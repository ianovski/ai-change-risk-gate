package risk

import (
	"testing"

	"github.com/ianovski/ai-change-risk-gate/internal/model"
)

func TestScore_HighRiskInfraNoTests(t *testing.T) {
	scorer := NewDefaultScorer()
	score, level, reasons := scorer.Score(model.EvaluateRiskRequest{
		FilesChanged: 40,
		LinesAdded:   1200,
		LinesDeleted: 500,
		ChangedPaths: []string{"infra/terraform/main.tf", "services/auth/main.go"},
		HasTests:     false,
	})
	if score < 65 {
		t.Fatalf("expected high score, got %d", score)
	}
	if level != model.RiskHigh && level != model.RiskCritical {
		t.Fatalf("expected high or critical level, got %s", level)
	}
	if len(reasons) == 0 {
		t.Fatal("expected reasons")
	}
}
