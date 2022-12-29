// Package flg defines flags, common for all commands.
package flg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Semior001/releaseit/app/git/engine"
)

// EngineGroup defines parameters for the engine.
type EngineGroup struct {
	Type   string      `long:"type" env:"TYPE" choice:"github" choice:"gitlab" description:"type of the repository engine" required:"true"`
	Github GithubGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Gitlab GitlabGroup `group:"gitlab" namespace:"gitlab" env-namespace:"GITLAB"`
}

// Build builds the engine.
func (r EngineGroup) Build() (engine.Interface, error) {
	switch r.Type {
	case "github":
		return engine.NewGithub(
			r.Github.Repo.Owner,
			r.Github.Repo.Name,
			r.Github.BasicAuth.Username,
			r.Github.BasicAuth.Password,
			http.Client{Timeout: r.Github.Timeout},
		)
	case "gitlab":
		return engine.NewGitlab(
			r.Gitlab.Token,
			r.Gitlab.BaseURL,
			r.Gitlab.ProjectID,
			http.Client{Timeout: r.Gitlab.Timeout},
		)
	}
	return nil, fmt.Errorf("unsupported repository engine type %s", r.Type)
}

// GithubGroup defines parameters to connect to the github repository.
type GithubGroup struct {
	Repo struct {
		Owner string `long:"owner" env:"OWNER" description:"owner of the repository"`
		Name  string `long:"name" env:"NAME" description:"name of the repository"`
	} `group:"repo" namespace:"repo" env-namespace:"REPO"`
	BasicAuth struct {
		Username string `long:"username" env:"USERNAME" description:"username for basic auth"`
		Password string `long:"password" env:"PASSWORD" description:"password for basic auth"`
	} `group:"basic_auth" namespace:"basic_auth" env-namespace:"BASIC_AUTH"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

// GitlabGroup defines parameters to connect to the gitlab repository.
type GitlabGroup struct {
	Token     string        `long:"token" env:"TOKEN" description:"token to connect to the gitlab repository"`
	BaseURL   string        `long:"base_url" env:"BASE_URL" description:"base url of the gitlab instance"`
	ProjectID string        `long:"project_id" env:"PROJECT_ID" description:"project id of the repository"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}
