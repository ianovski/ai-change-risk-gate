package policy

import (
	"testing"

	"github.com/ianovski/ai-change-risk-gate/internal/model"
)

func TestDecide_BlockVeryHigh(t *testing.T) {
	engine := NewDefaultEngine()
	decision := engine.Decide(model.EvaluateRiskRequest{}, 95, model.RiskCritical, []string{"very high risk"})
	if decision != model.DecisionBlock {
		t.Fatalf("expected block, got %s", decision)
	}
}

func TestDecide_RequireApprovalHigh(t *testing.T) {
	engine := NewDefaultEngine()
	decision := engine.Decide(model.EvaluateRiskRequest{HasTests: true}, 70, model.RiskHigh, []string{"high risk"})
	if decision != model.DecisionRequireApproval {
		t.Fatalf("expected require_approval, got %s", decision)
	}
}
