# ReleaseIt! [![Go](https://github.com/Semior001/releaseit/actions/workflows/.go.yaml/badge.svg)](https://github.com/Semior001/releaseit/actions/workflows/.go.yaml) [![go report card](https://goreportcard.com/badge/github.com/semior001/releaseit)](https://goreportcard.com/report/github.com/semior001/releaseit) [![PkgGoDev](https://pkg.go.dev/badge/github.com/Semior001/releaseit)](https://pkg.go.dev/github.com/Semior001/releaseit) [![codecov](https://codecov.io/gh/Semior001/releaseit/branch/master/graph/badge.svg?token=0MAV99RJ1C)](https://codecov.io/gh/Semior001/releaseit)

Utility for generating and publishing changelogs to different destinations.

Inspired by [mikepenz/release-changelog-builder-action](https://github.com/mikepenz/release-changelog-builder-action)

## All application options
```
Application Options:
      --dbg                                    turn on debug mode [$DEBUG]

Help Options:
  -h, --help                                   Show this help message

[release command options]
          --tag=                               tag to be released [$TAG]

    engine:
          --engine.type=[github]               type of the repository engine [$ENGINE_TYPE]

    repo:
          --engine.github.repo.owner=          owner of the repository [$ENGINE_GITHUB_REPO_OWNER]
          --engine.github.repo.name=           name of the repository [$ENGINE_GITHUB_REPO_NAME]

    basic_auth:
          --engine.github.basic_auth.username= username for basic auth [$ENGINE_GITHUB_BASIC_AUTH_USERNAME]
          --engine.github.basic_auth.password= password for basic auth [$ENGINE_GITHUB_BASIC_AUTH_PASSWORD]

    telegram:
          --notify.telegram.chat_id=           id of the chat, where the release notes will be sent [$NOTIFY_TELEGRAM_CHAT_ID]
          --notify.telegram.token=             bot token [$NOTIFY_TELEGRAM_TOKEN]
          --notify.telegram.web_page_preview   request telegram to preview for web links [$NOTIFY_TELEGRAM_WEB_PAGE_PREVIEW]
          --notify.telegram.conf_location=     location to the config file [$NOTIFY_TELEGRAM_CONF_LOCATION]

    github:
          --notify.github.release_name_tmpl=   template for release name [$NOTIFY_GITHUB_RELEASE_NAME_TMPL]
          --notify.github.conf_location=       location to the config file [$NOTIFY_GITHUB_CONF_LOCATION]

    repo:
          --notify.github.repo.owner=          owner of the repository [$NOTIFY_GITHUB_REPO_OWNER]
          --notify.github.repo.name=           name of the repository [$NOTIFY_GITHUB_REPO_NAME]

    basic_auth:
          --notify.github.basic_auth.username= username for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_USERNAME]
          --notify.github.basic_auth.password= password for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_PASSWORD]

    stdout:
          --notify.stdout.conf_location=       location to the config file [$NOTIFY_STDOUT_CONF_LOCATION]
```

## Release notes builder configuration
| Name              | Description                                                                                                                                             |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| categories        | Categories of pull requests                                                                                                                             |
| categories.title  | Title, which will be provided to the release notes template                                                                                             |
| categories.labels | An array of labels, to match pull request labels against. If any PR label matches any category label, the pull request will show up under this category |
| ignore_labels     | An array of labels, to match pull request labels against. If PR contains any of the defined ignore labels - this PR won't be provided to the template   |
| sort_field        | Field, by which pull requests must be sorted, in format +&#124;-field currently supported fields: `number`, `author`, `title`, `closed`                     |
| template          | Template for a changelog in golang's text template language                                                                                             |
| empty_template    | Template for release with no changes                                                                                                                    |
| unused_title      | If set, the unused category will be built under this title at the end of the changelog                                                                  |

## Example release notes builder configuration

```yaml
categories:
  - title: "## 🚀 Features"
    labels:
      - "feature"
  - title: "## 🐛 Fixes"
    labels:
      - "fix"
  - title: "## 🧰 Maintenance"
    labels:
      - "maintenance"
unused_title: "## ❓ Unlabeled"
ignore_labels:
  - "ignore"
sort_field: "-number"
template: |
  Project: Example release config
  Development area: Backend
  Version {{.Tag}}
  Date: {{.Date.Format "Jan 02, 2006 15:04:05 UTC"}}
  {{if not .Categories}}- No changes{{end}}{{range .Categories}}{{.Title}}
  {{range .PRs}}- {{.Title}} (#{{.Number}}) by @{{.Author}}
  {{end}}{{end}}
empty_template: "- no changes"
```

## Template variables for release notes builder

| Name                       | Description                                                      | Example                    |
|----------------------------|------------------------------------------------------------------|----------------------------|
| {{.Tag}}                   | Tag name of the release                                          | v1.0.0                     |
| {{.Date}}                  | Date of the commit which was tagged                              | Jan 02, 2006 15:04:05 UTC  |
|                            |                                                                  |                            |
| {{.Categories.Title}}      | Title of the category from the config                            | Features                   |
|                            |                                                                  |                            |
| {{.Categories.PRs.Number}} | Number of the pull request                                       | 642                        |
| {{.Categories.PRs.Title}}  | Title of the pull request                                        | Some awesome feature added |
| {{.Categories.PRs.Author}} | Username of the author of pull request                           | Semior001                  |
| {{.Categories.PRs.Closed}} | Timestamp, when the pull request was closed (might be empty)     | Jan 02, 2006 15:04:05 UTC  |

The golang's [text/template package](https://pkg.go.dev/text/template) is used for executing template for release notes.

## (Github) Template variables for release title

| Name         | Description             | Example |
|--------------|-------------------------|---------|
| {{.TagName}} | Tag name of the release | v1.0.0  |
