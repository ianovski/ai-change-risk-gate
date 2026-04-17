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
