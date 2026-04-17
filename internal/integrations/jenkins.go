package integrations

import "github.com/ianovski/ai-change-risk-gate/internal/model"

type JenkinsBuildEvent struct {
	JobName                  string   `json:"job_name"`
	Repo                     string   `json:"repo"`
	Branch                   string   `json:"branch"`
	CommitSHA                string   `json:"commit_sha"`
	ChangedPaths             []string `json:"changed_paths"`
	FilesChanged             int      `json:"files_changed"`
	LinesAdded               int      `json:"lines_added"`
	LinesDeleted             int      `json:"lines_deleted"`
	HasTests                 bool     `json:"has_tests"`
	HasIntegrationTests      bool     `json:"has_integration_tests"`
	HasRollbackPlan          bool     `json:"has_rollback_plan"`
	AuthorFailureRate        float64  `json:"author_recent_failure_rate"`
	DeploymentWindowCritical bool     `json:"deployment_window_critical"`
}

func ConvertJenkinsBuild(e JenkinsBuildEvent) model.EvaluateRiskRequest {
	return model.EvaluateRiskRequest{
		Repo:                     e.Repo,
		Branch:                   e.Branch,
		CommitSHA:                e.CommitSHA,
		FilesChanged:             e.FilesChanged,
		LinesAdded:               e.LinesAdded,
		LinesDeleted:             e.LinesDeleted,
		ChangedPaths:             e.ChangedPaths,
		HasRollbackPlan:          e.HasRollbackPlan,
		HasTests:                 e.HasTests,
		HasIntegrationTests:      e.HasIntegrationTests,
		AuthorRecentFailureRate:  e.AuthorFailureRate,
		DeploymentWindowCritical: e.DeploymentWindowCritical,
	}
}
