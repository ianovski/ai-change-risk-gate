package integrations

import "github.com/ianovski/ai-change-risk-gate/internal/model"

type GitHubPullRequestEvent struct {
	Repository string   `json:"repository"`
	Branch     string   `json:"branch"`
	CommitSHA  string   `json:"commit_sha"`
	Files      int      `json:"files"`
	Additions  int      `json:"additions"`
	Deletions  int      `json:"deletions"`
	Paths      []string `json:"paths"`
	HasTests   bool     `json:"has_tests"`
	HasITests  bool     `json:"has_integration_tests"`
}

func ConvertGitHubPR(e GitHubPullRequestEvent) model.EvaluateRiskRequest {
	return model.EvaluateRiskRequest{
		Repo:                e.Repository,
		Branch:              e.Branch,
		CommitSHA:           e.CommitSHA,
		FilesChanged:        e.Files,
		LinesAdded:          e.Additions,
		LinesDeleted:        e.Deletions,
		ChangedPaths:        e.Paths,
		HasRollbackPlan:     false,
		HasTests:            e.HasTests,
		HasIntegrationTests: e.HasITests,
	}
}
