package policy

import (
	"strings"

	"github.com/ianovski/ai-change-risk-gate/internal/model"
)

type Engine interface {
	Decide(req model.EvaluateRiskRequest, score int, level model.RiskLevel, reasons []string) model.Decision
}

type DefaultEngine struct{}

func NewDefaultEngine() *DefaultEngine {
	return &DefaultEngine{}
}

func (e *DefaultEngine) Decide(req model.EvaluateRiskRequest, score int, level model.RiskLevel, reasons []string) model.Decision {
	if score >= 92 {
		return model.DecisionBlock
	}

	hasSecurityReason := false
	hasDatabaseReason := false
	for _, r := range reasons {
		lr := strings.ToLower(r)
		if strings.Contains(lr, "security") {
			hasSecurityReason = true
		}
		if strings.Contains(lr, "database") {
			hasDatabaseReason = true
		}
	}

	if hasSecurityReason && !req.HasTests {
		return model.DecisionBlock
	}

	if hasDatabaseReason && !req.HasRollbackPlan {
		return model.DecisionRequireApproval
	}

	if level == model.RiskHigh || level == model.RiskCritical {
		return model.DecisionRequireApproval
	}

	return model.DecisionAllow
}
