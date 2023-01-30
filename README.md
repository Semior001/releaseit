# ReleaseIt! [![Go](https://github.com/Semior001/releaseit/actions/workflows/.go.yaml/badge.svg)](https://github.com/Semior001/releaseit/actions/workflows/.go.yaml) [![go report card](https://goreportcard.com/badge/github.com/semior001/releaseit)](https://goreportcard.com/report/github.com/semior001/releaseit) [![PkgGoDev](https://pkg.go.dev/badge/github.com/Semior001/releaseit)](https://pkg.go.dev/github.com/Semior001/releaseit) [![codecov](https://codecov.io/gh/Semior001/releaseit/branch/master/graph/badge.svg?token=0MAV99RJ1C)](https://codecov.io/gh/Semior001/releaseit)

Utility for generating and publishing changelogs to different destinations.

Inspired by [mikepenz/release-changelog-builder-action](https://github.com/mikepenz/release-changelog-builder-action)

## All application options
<details>
<summary>Click to expand</summary>

```
Application Options:
      --dbg                                    turn on debug mode [$DEBUG]

Help Options:
  -h, --help                                   Show this help message

[preview command options]
          --data-file=     path to the file with release data [$DATA_FILE]
          --extras=        extra variables to use in the template, will be merged (env primary) with ones in the config file [$EXTRAS]
          --conf_location= location to the config file [$CONF_LOCATION]

[changelog command options]
          --from=                              sha to start release notes from (default: {{ previous_tag .To }}) [$FROM]
          --to=                                sha to end release notes to (default: {{ last_tag }}) [$TO]
          --timeout=                           timeout for assembling the release (default: 5m) [$TIMEOUT]
          --squash-commit-rx=                  regexp to match squash commits (default: ^squash:(.?)+$) [$SQUASH_COMMIT_RX]
          --conf_location=                     location to the config file [$CONF_LOCATION]
          --extras=                            extra variables to use in the template [$EXTRAS]

    engine:
          --engine.type=[github|gitlab]        type of the repository engine [$ENGINE_TYPE]

    github:
          --engine.github.timeout=             timeout for http requests (default: 5s) [$ENGINE_GITHUB_TIMEOUT]

    repo:
          --engine.github.repo.owner=          owner of the repository [$ENGINE_GITHUB_REPO_OWNER]
          --engine.github.repo.name=           name of the repository [$ENGINE_GITHUB_REPO_NAME]

    basic_auth:
          --engine.github.basic_auth.username= username for basic auth [$ENGINE_GITHUB_BASIC_AUTH_USERNAME]
          --engine.github.basic_auth.password= password for basic auth [$ENGINE_GITHUB_BASIC_AUTH_PASSWORD]

    gitlab:
          --engine.gitlab.token=               token to connect to the gitlab repository [$ENGINE_GITLAB_TOKEN]
          --engine.gitlab.base_url=            base url of the gitlab instance [$ENGINE_GITLAB_BASE_URL]
          --engine.gitlab.project_id=          project id of the repository [$ENGINE_GITLAB_PROJECT_ID]
          --engine.gitlab.timeout=             timeout for http requests (default: 5s) [$ENGINE_GITLAB_TIMEOUT]

    notify:
          --notify.stdout                      print release notes to stdout [$NOTIFY_STDOUT]

    telegram:
          --notify.telegram.chat_id=           id of the chat, where the release notes will be sent [$NOTIFY_TELEGRAM_CHAT_ID]
          --notify.telegram.token=             bot token [$NOTIFY_TELEGRAM_TOKEN]
          --notify.telegram.web_page_preview   request telegram to preview for web links [$NOTIFY_TELEGRAM_WEB_PAGE_PREVIEW]
          --notify.telegram.timeout=           timeout for http requests (default: 5s) [$NOTIFY_TELEGRAM_TIMEOUT]

    github:
          --notify.github.timeout=             timeout for http requests (default: 5s) [$NOTIFY_GITHUB_TIMEOUT]
          --notify.github.release_name_tmpl=   template for release name [$NOTIFY_GITHUB_RELEASE_NAME_TMPL]

    repo:
          --notify.github.repo.owner=          owner of the repository [$NOTIFY_GITHUB_REPO_OWNER]
          --notify.github.repo.name=           name of the repository [$NOTIFY_GITHUB_REPO_NAME]

    basic_auth:
          --notify.github.basic_auth.username= username for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_USERNAME]
          --notify.github.basic_auth.password= password for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_PASSWORD]

    mattermost-hook:
          --notify.mattermost-hook.base_url=   base url of the mattermost server [$NOTIFY_MATTERMOST_HOOK_BASE_URL]
          --notify.mattermost-hook.id=         id of the hook, where the release notes will be sent [$NOTIFY_MATTERMOST_HOOK_ID]
          --notify.mattermost-hook.timeout=    timeout for http requests (default: 5s) [$NOTIFY_MATTERMOST_HOOK_TIMEOUT]

    post:
          --notify.post.url=                   url to send the release notes [$NOTIFY_POST_URL]
          --notify.post.timeout=               timeout for http requests (default: 5s) [$NOTIFY_POST_TIMEOUT]
```

</details>

**Note**: `from` and `to` options of `changelog` command accept expressions, which must be written in gotemplate manner.
`from` expression can relate on the value of `to`, `to` is evaluated first.

Example (from .env file): `TO='${{ last_commit "develop" }}'`

Supported functions:
- `last_commit(branch_name)`
- `previous_tag(tag_name)`
- `tags()` - returns list of tags in descending order
- `last_tag()` - returns the last tag in repository (shortcut for `{{ index (tags) 0 }}`)

## Preview data file structure
```go
var data struct {
    Version      string            `yaml:"version"`
    FromSHA      string            `yaml:"from_sha"`
    ToSHA        string            `yaml:"to_sha"`
    Extras       map[string]string `yaml:"extras"`
    PullRequests []git.PullRequest `yaml:"pull_requests"`
}

type PullRequest struct {
    Number   int       `yaml:"number"`
    Title    string    `yaml:"title"`
    Body     string    `yaml:"body"`
    Author   User      `yaml:"author"`
    Labels   []string  `yaml:"labels"`
    ClosedAt time.Time `yaml:"closed_at"`
    Branch   string    `yaml:"branch"`
    URL      string    `yaml:"url"`
}

type User struct {
    Date     time.Time `yaml:"date"`
    Username string    `yaml:"username"`
    Email    string    `yaml:"email"`
}
```

See [example](_example/preview_data.yaml) for details.

## Release notes builder configuration
```go
// Config describes the configuration of the changelog builder.
type Config struct {
	// categories to parse in pull requests
	Categories []CategoryConfig `yaml:"categories"`

	// field, by which pull requests must be sorted, in format +|-field
	// currently supported fields: number, author, title, closed
	SortField string `yaml:"sort_field"`
	// template for a changelog.
	Template string `yaml:"template"`

	// if set, the unused category will be built under this title at the
	// end of the changelog
	UnusedTitle string `yaml:"unused_title"`
	// labels for pull requests, which won't be in release notes
	IgnoreLabels []string `yaml:"ignore_labels"`
	// regexp for pull request branches, which won't be in release notes
	IgnoreBranch string `yaml:"ignore_branch"`
	// compiled regexp, used internally
	IgnoreBranchRe *regexp.Regexp `yaml:"-"`
}

// CategoryConfig describes the category configuration.
type CategoryConfig struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`

	// regexp to match branch name
	Branch string `yaml:"branch"`

	// compiled branch regexp, used internally
	BranchRe *regexp.Regexp `yaml:"-"`
}
```

## Template variables for release notes builder
```go
type tmplData struct {
	FromSHA    string
	ToSHA      string
	Categories []categoryTmplData
	Date       time.Time // always set to the time when the changelog is generated
	Extras     map[string]string
}

type categoryTmplData struct {
	Title string
	PRs   []prTmplData
}

type prTmplData struct {
	Number   int
	Title    string
	Author   string
	URL      string
	Branch   string
	ClosedAt time.Time
}
```

The golang's [text/template package](https://pkg.go.dev/text/template) is used for executing template for release notes. 
It also imports functions from [sprig](http://masterminds.github.io/sprig/) (excluding `env` and `expandenv`) library in 
order to provide common used template functions.

## (Github) Template variables for release title

| Name         | Description             | Example |
|--------------|-------------------------|---------|
| {{.TagName}} | Tag name of the release | v1.0.0  |

See [example](_example/config.yaml) for more details.
Sprig functions are also available.
