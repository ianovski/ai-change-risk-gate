package risk

import (
	"math"
	"strings"

	"github.com/ianovski/ai-change-risk-gate/internal/model"
)

type Scorer interface {
	Score(req model.EvaluateRiskRequest) (score int, level model.RiskLevel, reasons []string)
}

type DefaultScorer struct{}

func NewDefaultScorer() *DefaultScorer {
	return &DefaultScorer{}
}

func (s *DefaultScorer) Score(req model.EvaluateRiskRequest) (int, model.RiskLevel, []string) {
	raw := 0.0
	reasons := make([]string, 0, 8)

	if req.FilesChanged > 50 {
		raw += 18
		reasons = append(reasons, "large file change set")
	} else if req.FilesChanged > 20 {
		raw += 10
		reasons = append(reasons, "medium file change set")
	}

	churn := req.LinesAdded + req.LinesDeleted
	if churn > 2500 {
		raw += 18
		reasons = append(reasons, "very high code churn")
	} else if churn > 1000 {
		raw += 12
		reasons = append(reasons, "high code churn")
	} else if churn > 300 {
		raw += 6
		reasons = append(reasons, "moderate code churn")
	}

	hasInfra := containsPathFragment(req.ChangedPaths, []string{"terraform", "helm", "k8s", "cloudformation", "pulumi"})
	hasSecurity := containsPathFragment(req.ChangedPaths, []string{"auth", "iam", "secrets", "policy", "security"})
	hasDatabase := containsPathFragment(req.ChangedPaths, []string{"migration", "schema", "database", "sql"})

	if hasInfra {
		raw += 20
		reasons = append(reasons, "infrastructure changes detected")
	}
	if hasSecurity {
		raw += 20
		reasons = append(reasons, "security-sensitive changes detected")
	}
	if hasDatabase {
		raw += 14
		reasons = append(reasons, "database changes detected")
	}

	if !req.HasTests {
		raw += 18
		reasons = append(reasons, "no tests provided")
	} else if !req.HasIntegrationTests {
		raw += 6
		reasons = append(reasons, "no integration tests provided")
	}

	if !req.HasRollbackPlan {
		raw += 12
		reasons = append(reasons, "no rollback plan")
	}

	if req.AuthorRecentFailureRate >= 0.30 {
		raw += 10
		reasons = append(reasons, "author failure-rate signal elevated")
	}

	if req.DeploymentWindowCritical {
		raw += 10
		reasons = append(reasons, "critical deployment window")
	}

	score := int(math.Round(math.Min(math.Max(raw, 0), 100)))
	level := toRiskLevel(score)
	if len(reasons) == 0 {
		reasons = append(reasons, "low-risk profile")
	}
	return score, level, reasons
}

func toRiskLevel(score int) model.RiskLevel {
	switch {
	case score >= 85:
		return model.RiskCritical
	case score >= 65:
		return model.RiskHigh
	case score >= 35:
		return model.RiskMedium
	default:
		return model.RiskLow
	}
}

func containsPathFragment(paths []string, fragments []string) bool {
	for _, p := range paths {
		lp := strings.ToLower(p)
		for _, f := range fragments {
			if strings.Contains(lp, f) {
				return true
			}
		}
	}
	return false
}
