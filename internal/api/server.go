package api

import (
	"net/http"

	"github.com/ianovski/ai-change-risk-gate/internal/policy"
	"github.com/ianovski/ai-change-risk-gate/internal/risk"
	"github.com/ianovski/ai-change-risk-gate/internal/store"
)

type Server struct {
	mux      *http.ServeMux
	handlers *Handlers
}

func NewServer(scorer risk.Scorer, engine policy.Engine, approvals store.ApprovalStore) http.Handler {
	h := &Handlers{
		scorer: scorer,
		engine: engine,
		store:  approvals,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", withMethod(http.MethodGet, h.health))
	mux.HandleFunc("/v1/risk/evaluate", withMethod(http.MethodPost, h.evaluateRisk))
	mux.HandleFunc("/v1/approvals/override", withMethod(http.MethodPost, h.overrideApproval))
	mux.HandleFunc("/v1/approvals/", withMethod(http.MethodGet, h.getApproval))
	mux.HandleFunc("/v1/webhooks/github/pull_request", withMethod(http.MethodPost, h.githubPullRequest))
	mux.HandleFunc("/v1/webhooks/jenkins/build", withMethod(http.MethodPost, h.jenkinsBuild))
	return mux
}

func withMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}
