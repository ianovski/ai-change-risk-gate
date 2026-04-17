package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ianovski/ai-change-risk-gate/internal/integrations"
	"github.com/ianovski/ai-change-risk-gate/internal/model"
	"github.com/ianovski/ai-change-risk-gate/internal/policy"
	"github.com/ianovski/ai-change-risk-gate/internal/risk"
	"github.com/ianovski/ai-change-risk-gate/internal/store"
)

type Handlers struct {
	scorer risk.Scorer
	engine policy.Engine
	store  store.ApprovalStore
}

func (h *Handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handlers) evaluateRisk(w http.ResponseWriter, r *http.Request) {
	var req model.EvaluateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	resp := h.evaluate(req)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) githubPullRequest(w http.ResponseWriter, r *http.Request) {
	var event integrations.GitHubPullRequestEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid GitHub payload")
		return
	}
	resp := h.evaluate(integrations.ConvertGitHubPR(event))
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) jenkinsBuild(w http.ResponseWriter, r *http.Request) {
	var event integrations.JenkinsBuildEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid Jenkins payload")
		return
	}
	resp := h.evaluate(integrations.ConvertJenkinsBuild(event))
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) evaluate(req model.EvaluateRiskRequest) model.EvaluateRiskResponse {
	score, level, reasons := h.scorer.Score(req)
	suggested := h.engine.Decide(req, score, level, reasons)
	resp := model.EvaluateRiskResponse{
		RiskScore:         score,
		RiskLevel:         level,
		Reasons:           reasons,
		SuggestedDecision: suggested,
		EffectiveDecision: suggested,
	}

	if suggested != model.DecisionAllow {
		approvalID := newID()
		record := model.ApprovalRecord{
			ID:                approvalID,
			Repo:              req.Repo,
			CommitSHA:         req.CommitSHA,
			SuggestedDecision: suggested,
			EffectiveDecision: suggested,
			Status:            "pending",
			Reason:            strings.Join(reasons, "; "),
		}
		h.store.Put(record)
		resp.ApprovalID = approvalID
	}

	return resp
}

func (h *Handlers) overrideApproval(w http.ResponseWriter, r *http.Request) {
	var req model.ApprovalOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(req.ApprovalID) == "" || strings.TrimSpace(req.Approver) == "" {
		writeError(w, http.StatusBadRequest, "approval_id and approver are required")
		return
	}
	if len(strings.TrimSpace(req.Justification)) < 20 {
		writeError(w, http.StatusBadRequest, "justification must be at least 20 characters")
		return
	}

	rec, err := h.store.Get(req.ApprovalID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "approval not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if rec.Status == "approved" {
		writeJSON(w, http.StatusOK, rec)
		return
	}
	if rec.SuggestedDecision == model.DecisionBlock && !req.BreakGlass {
		writeError(w, http.StatusForbidden, "break_glass=true required to override blocked decisions")
		return
	}

	rec.Status = "approved"
	rec.Approver = req.Approver
	rec.Justification = req.Justification
	rec.EffectiveDecision = model.DecisionAllow
	if err := h.store.Update(rec); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update approval")
		return
	}

	writeJSON(w, http.StatusOK, rec)
}

func (h *Handlers) getApproval(w http.ResponseWriter, r *http.Request) {
	prefix := "/v1/approvals/"
	id := strings.TrimPrefix(r.URL.Path, prefix)
	if id == "" || id == r.URL.Path {
		writeError(w, http.StatusBadRequest, "approval id missing")
		return
	}
	rec, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "approval not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}

func newID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "fallback-id"
	}
	return hex.EncodeToString(buf)
}
