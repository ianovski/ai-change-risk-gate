package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ianovski/ai-change-risk-gate/internal/api"
	"github.com/ianovski/ai-change-risk-gate/internal/policy"
	"github.com/ianovski/ai-change-risk-gate/internal/risk"
	"github.com/ianovski/ai-change-risk-gate/internal/store"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	scorer := risk.NewDefaultScorer()
	engine := policy.NewDefaultEngine()
	approvals := store.NewMemoryApprovalStore()

	handler := api.NewServer(scorer, engine, approvals)

	log.Printf("ai-change-risk-gate listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
