# ReleaseIt!

Utility for generating and publishing changelogs to different destinations.

Inspired by [mikepenz/release-changelog-builder-action](https://github.com/mikepenz/release-changelog-builder-action)

## All application options
```
Application Options:
      --dbg                             turn on debug mode [$DEBUG]

Help Options:
  -h, --help                            Show this help message

[release command options]
          --conf_location=              location to the config file
                                        [$CONF_LOCATION]

    repo:
          --github.repo.owner=          owner of the repository
                                        [$GITHUB_REPO_OWNER]
          --github.repo.name=           name of the repository
                                        [$GITHUB_REPO_NAME]

    basic_auth:
          --github.basic_auth.username= username for basic auth
                                        [$GITHUB_BASIC_AUTH_USERNAME]
          --github.basic_auth.password= password for basic auth
                                        [$GITHUB_BASIC_AUTH_PASSWORD]
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