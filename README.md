# AI Change Risk Gate

Deterministic risk-gating service for CI/CD workflows with policy fallback and explicit override controls.

## Why

This service acts like a deployment risk gate for pull requests and builds:

- Scores a change set with deterministic heuristics.
- Applies policy to allow, require approval, or block.
- Supports explicit override flow with break-glass protection for blocked changes.
- Includes provider adapters for GitHub PR and Jenkins build payloads.

## Run

```bash
go run ./cmd/server
```

Server defaults to `:8080`.

## Container

Build and run locally with Docker:

```bash
docker build -t ai-change-risk-gate .
docker run --rm -p 8080:8080 ai-change-risk-gate
```

## Endpoints

- `GET /healthz`
- `POST /v1/risk/evaluate`
- `POST /v1/approvals/override`
- `GET /v1/approvals/{id}`
- `POST /v1/webhooks/github/pull_request`
- `POST /v1/webhooks/jenkins/build`

## Example risk evaluation

```bash
curl -sS -X POST localhost:8080/v1/risk/evaluate \
  -H 'content-type: application/json' \
  -d '{
    "repo":"payments-api",
    "branch":"main",
    "commit_sha":"abc123",
    "files_changed":28,
    "lines_added":1400,
    "lines_deleted":240,
    "changed_paths":["infra/terraform/prod.tf","db/migrations/20260417_add_index.sql"],
    "has_rollback_plan":false,
    "has_tests":true,
    "has_integration_tests":false,
    "author_recent_failure_rate":0.33,
    "deployment_window_critical":true
  }'
```

If the response includes `approval_id`, query approval status and override it:

```bash
curl -sS localhost:8080/v1/approvals/<approval_id>

curl -sS -X POST localhost:8080/v1/approvals/override \
  -H 'content-type: application/json' \
  -d '{
    "approval_id":"<approval_id>",
    "approver":"oncall-sre",
    "justification":"Emergency fix for customer-facing outage. Rollback verified and stakeholder approval recorded.",
    "break_glass":true
  }'
```

## Notes

- This MVP uses in-memory approval storage.
- Next step is adding persistent store (Postgres), authn/authz, and signed webhook verification.

## Deployment model
Github Action workflow and private Cloud Run service

### Security model

Use these controls together:

1. Keep Cloud Run unauthenticated access disabled.
2. Create a dedicated Google service account with only `roles/run.invoker` on this service.
3. Let GitHub Actions impersonate that service account through Workload Identity Federation.
4. Use a `pull_request_target` workflow that does not check out or execute PR code.

This matters for a public repo: forked pull requests cannot be trusted with secrets, and `pull_request_target` is only safe here because the workflow reads PR metadata via the GitHub API and never runs code from the PR branch.

### Required GitHub variables

Configure these repository variables before enabling the workflow:

- `RISK_GATE_URL`: base HTTPS URL of the Cloud Run service, for example `https://ai-change-risk-gate-abc-uc.a.run.app`
- `GCP_WORKLOAD_IDENTITY_PROVIDER`: full provider resource name, for example `projects/123456789/locations/global/workloadIdentityPools/github/providers/actions`
- `GCP_SERVICE_ACCOUNT`: deploy-time invoker identity, for example `risk-gate-invoker@my-project.iam.gserviceaccount.com`
- `RISK_GATE_FAIL_ON`: optional comma-separated decisions that should fail the check. Default is `block,require_approval`.

### Cloud Run setup outline

Build and deploy:

```bash
gcloud builds submit --tag us-central1-docker.pkg.dev/PROJECT_ID/risk-gate/ai-change-risk-gate

gcloud run deploy ai-change-risk-gate \
  --image us-central1-docker.pkg.dev/PROJECT_ID/risk-gate/ai-change-risk-gate \
  --region us-central1 \
  --no-allow-unauthenticated
```

Create an invoker service account and grant only invoke access:

```bash
gcloud iam service-accounts create risk-gate-invoker

gcloud run services add-iam-policy-binding ai-change-risk-gate \
  --region us-central1 \
  --member serviceAccount:risk-gate-invoker@PROJECT_ID.iam.gserviceaccount.com \
  --role roles/run.invoker
```

Create a workload identity pool and provider for GitHub Actions. The provider should map repository claims and restrict access to this repository. Example attribute condition:

```text
assertion.repository == 'OWNER/REPO'
```

If you want tighter scope, also pin the workflow ref or environment in the provider condition.

### Workflow behavior

The included workflow in `.github/workflows/pr-risk-gate.yml`:

- triggers on `pull_request_target`
- avoids `actions/checkout`
- reads changed files from the GitHub API
- builds the normalized JSON payload expected by `/v1/risk/evaluate`
- authenticates to Google Cloud with OIDC
- calls the private Cloud Run service with an ID token
- comments the result back onto the PR
- fails the check when the configured decision requires it

### Current limitations

- Approval state is not durable because records are stored in memory.
- `require_approval` currently only produces an approval record and cannot yet map to a real GitHub approval workflow.
- The GitHub-specific webhook endpoint is still a custom payload endpoint, not a verified native GitHub webhook receiver.

For the next production step, add a persistent approval store and either a GitHub App integration or native webhook signature verification.
