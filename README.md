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
          --data-file=                         path to the file with release data [$DATA_FILE]
          --conf_location=                     location to the config file [$CONF_LOCATION]

[changelog command options]
          --from=                              sha to start release notes from [$FROM]
          --to=                                sha to end release notes to [$TO]

[release command options]
          --tag=                               tag to be released [$TAG]

[changelog & release options]
     --timeout=                                timeout for assembling the release (default: 5m) [$TIMEOUT]
     
     --squash-commit-rx=                       regexp to match squash commits (default: ^squash:(.?)+$) [$SQUASH_COMMIT_RX]

    engine:
          --engine.type=[github|gitlab]        type of the repository engine [$ENGINE_TYPE]

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

    notify:
          --notify.stdout                      print release notes to stdout [$NOTIFY_STDOUT]
          --notify.conf_location=              location to the config file [$NOTIFY_CONF_LOCATION]

    telegram:
          --notify.telegram.chat_id=           id of the chat, where the release notes will be sent [$NOTIFY_TELEGRAM_CHAT_ID]
          --notify.telegram.token=             bot token [$NOTIFY_TELEGRAM_TOKEN]
          --notify.telegram.web_page_preview   request telegram to preview for web links [$NOTIFY_TELEGRAM_WEB_PAGE_PREVIEW]

    github:
          --notify.github.release_name_tmpl=   template for release name [$NOTIFY_GITHUB_RELEASE_NAME_TMPL]

    repo:
          --notify.github.repo.owner=          owner of the repository [$NOTIFY_GITHUB_REPO_OWNER]
          --notify.github.repo.name=           name of the repository [$NOTIFY_GITHUB_REPO_NAME]

    basic_auth:
          --notify.github.basic_auth.username= username for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_USERNAME]
          --notify.github.basic_auth.password= password for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_PASSWORD]

    mattermost:
          --notify.mattermost.base_url=        base url of the mattermost server [$NOTIFY_MATTERMOST_BASE_URL]
          --notify.mattermost.channel_id=      id of the channel, where the release notes will be sent [$NOTIFY_MATTERMOST_CHANNEL_ID]
          --notify.mattermost.login_id=        login id of the user, who will send the release notes [$NOTIFY_MATTERMOST_LOGIN_ID]
          --notify.mattermost.password=        password of the user, who will send the release notes [$NOTIFY_MATTERMOST_PASSWORD]
          --notify.mattermost.ldap             use ldap auth [$NOTIFY_MATTERMOST_LDAP]

    mattermost-hook:
          --notify.mattermost-hook.base_url=   base url of the mattermost server [$NOTIFY_MATTERMOST_HOOK_BASE_URL]
          --notify.mattermost-hook.id=         id of the hook, where the release notes will be sent [$NOTIFY_MATTERMOST_HOOK_ID]
```

</details>

**Note**: `from` and `to` options of `changelog` command accept expressions, which must be written in gotemplate manner,
and the whole expressions should start from `!!` prefix.

Example (from .env file): `TO='${{ last_commit "develop" }}'`

Supported functions:
- `last_commit(branch_name)`
- `head`

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
| Name              | Description                                                                                                                                             |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| categories        | Categories of pull requests                                                                                                                             |
| categories.title  | Title, which will be provided to the release notes template                                                                                             |
| categories.labels | An array of labels, to match pull request labels against. If any PR label matches any category label, the pull request will show up under this category |
| categories.branch | A regular expression to match branch name to the corresponding category.                                                                                |
| ignore_labels     | An array of labels, to match pull request labels against. If PR contains any of the defined ignore labels - this PR won't be provided to the template   |
| sort_field        | Field, by which pull requests must be sorted, in format +&#124;-field currently supported fields: `number`, `author`, `title`, `closed`                 |
| template          | Template for a changelog in golang's text template language                                                                                             |
| empty_template    | Template for release with no changes                                                                                                                    |
| unused_title      | If set, the unused category will be built under this title at the end of the changelog                                                                  |

## Template variables for release notes builder

| Name                         | Description                                                  | Example                                         |
|------------------------------|--------------------------------------------------------------|-------------------------------------------------|
| {{.Version}}                 | Version name of the release, might be tag or diff.           | `v1.0.0` or `2bda1d3...82e35cf`                 |
| {{.FromSHA }}                | SHA or tag where diff starts at                              | `v1.0.0` or `2bda1d3`                           |
| {{.ToSHA }}                  | SHA or tag where diff ends at                                | `v2.0.0` or `82e35cf`                           |
| {{.Date}}                    | Date of the commit which was tagged                          | Jan 02, 2006 15:04:05 UTC                       |
|                              |                                                              |                                                 |
| {{.Categories.Title}}        | Title of the category from the config                        | Features                                        |
|                              |                                                              |                                                 |
| {{.Categories.PRs.Number}}   | Number of the pull request                                   | 642                                             |
| {{.Categories.PRs.Title}}    | Title of the pull request                                    | Some awesome feature added                      |
| {{.Categories.PRs.Author}}   | Username of the author of pull request                       | Semior001                                       |
| {{.Categories.PRs.ClosedAt}} | Timestamp, when the pull request was closed (might be empty) | Jan 02, 2006 15:04:05 UTC                       |
| {{.Categories.PRs.URL}}      | URL to the pull request                                      | `https://github.com/Semior001/releaseit/pull/6` |

The golang's [text/template package](https://pkg.go.dev/text/template) is used for executing template for release notes. 
It also imports functions from [sprig](http://masterminds.github.io/sprig/) (excluding `env` and `expandenv`) library in 
order to provide common used template functions.

## (Github) Template variables for release title

| Name         | Description             | Example |
|--------------|-------------------------|---------|
| {{.TagName}} | Tag name of the release | v1.0.0  |

See [example](_example/config.yaml) for more details.
