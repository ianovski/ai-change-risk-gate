package model

type EvaluateRiskRequest struct {
	Repo                     string   `json:"repo"`
	Branch                   string   `json:"branch"`
	CommitSHA                string   `json:"commit_sha"`
	FilesChanged             int      `json:"files_changed"`
	LinesAdded               int      `json:"lines_added"`
	LinesDeleted             int      `json:"lines_deleted"`
	ChangedPaths             []string `json:"changed_paths"`
	HasRollbackPlan          bool     `json:"has_rollback_plan"`
	HasTests                 bool     `json:"has_tests"`
	HasIntegrationTests      bool     `json:"has_integration_tests"`
	AuthorRecentFailureRate  float64  `json:"author_recent_failure_rate"`
	DeploymentWindowCritical bool     `json:"deployment_window_critical"`
}

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type Decision string

const (
	DecisionAllow           Decision = "allow"
	DecisionRequireApproval Decision = "require_approval"
	DecisionBlock           Decision = "block"
)

type EvaluateRiskResponse struct {
	RiskScore         int       `json:"risk_score"`
	RiskLevel         RiskLevel `json:"risk_level"`
	Reasons           []string  `json:"reasons"`
	SuggestedDecision Decision  `json:"suggested_decision"`
	EffectiveDecision Decision  `json:"effective_decision"`
	ApprovalID        string    `json:"approval_id,omitempty"`
}

type ApprovalRecord struct {
	ID                string   `json:"id"`
	Repo              string   `json:"repo"`
	CommitSHA         string   `json:"commit_sha"`
	SuggestedDecision Decision `json:"suggested_decision"`
	EffectiveDecision Decision `json:"effective_decision"`
	Status            string   `json:"status"`
	Reason            string   `json:"reason"`
	Approver          string   `json:"approver,omitempty"`
	Justification     string   `json:"justification,omitempty"`
}

type ApprovalOverrideRequest struct {
	ApprovalID    string `json:"approval_id"`
	Approver      string `json:"approver"`
	Justification string `json:"justification"`
	BreakGlass    bool   `json:"break_glass"`
}

// Provider-neutral payload from provider adapters.
type RiskInputEnvelope struct {
	Provider  string              `json:"provider"`
	EventType string              `json:"event_type"`
	Payload   EvaluateRiskRequest `json:"payload"`
}
