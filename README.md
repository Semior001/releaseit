# ReleaseIt! [![Go](https://github.com/Semior001/releaseit/actions/workflows/.go.yaml/badge.svg)](https://github.com/Semior001/releaseit/actions/workflows/.go.yaml) [![go report card](https://goreportcard.com/badge/github.com/semior001/releaseit)](https://goreportcard.com/report/github.com/semior001/releaseit) [![PkgGoDev](https://pkg.go.dev/badge/github.com/Semior001/releaseit)](https://pkg.go.dev/github.com/Semior001/releaseit) [![codecov](https://codecov.io/gh/Semior001/releaseit/branch/master/graph/badge.svg?token=0MAV99RJ1C)](https://codecov.io/gh/Semior001/releaseit)

Utility for generating and publishing changelogs to different destinations.

Inspired by [mikepenz/release-changelog-builder-action](https://github.com/mikepenz/release-changelog-builder-action)

## Installation
ReleaseIt is distributed as a docker image. You can pull it from [ghcr package](https://github.com/Semior001/releaseit/pkgs/container/releaseit).

Env vars configuration example is available [here](_example/simple-prs/example.env).

## How it works.
It looks for the closed pull requests attached to merge or squash commits between the references provided in arguments,
and splits them into categories, which are defined in the configuration file.

A commit with two or more parents is considered a merge commit. Squash commits are matched by the regexp
provided in arguments (default is `^.*#\d+.*$` - any commit message with `#` and a number in it).

To avoid repeats of pull requests among the releases (e.g., in case a developer resolved conflicts by merging the main
branch into the feature branch, so the merge commit will be repeated in the commit history); it checks that the
merge/squash commit is after the commit provided within the `FROM` argument.

## All application options
<details>
<summary>Click to expand</summary>

```
Application Options:
      --dbg                turn on debug mode [$DEBUG]

Help Options:
  -h, --help               Show this help message

[preview command options]
          --data-file=     path to the file with release data [$DATA_FILE]
          --extras=        extra variables to use in the template, will be merged (env primary) with ones in the config file [$EXTRAS]
          --conf-location= location to the config file [$CONF_LOCATION]

[changelog command options]
          --from=                              commit ref to start release notes from (default: {{ previousTag .To (headed (filter semver tags)) }}) [$FROM]
          --to=                                commit ref to end release notes to (default: {{ last (filter semver tags) }}) [$TO]
          --timeout=                           timeout for assembling the release (default: 5m) [$TIMEOUT]
          --fetch-merge-commits-filter=        regexp to filter merge commits (default: .*) [$FETCH_MERGE_COMMITS_FILTER]
          --conf-location=                     location to the config file [$CONF_LOCATION]
          --extras=                            extra variables to use in the template [$EXTRAS]
          --max-concurrent-pr-requests=        maximum number of concurrent PR requests (default: 10) [$MAX_CONCURRENT_PR_REQUESTS]

    engine:
          --engine.type=[github|gitlab]        type of the repository engine [$ENGINE_TYPE]

    github:
          --engine.github.timeout=             timeout for http requests (default: 5s) [$ENGINE_GITHUB_TIMEOUT]

    repo:
          --engine.github.repo.full-name=      full name of the repository (owner/name) [$ENGINE_GITHUB_REPO_FULL_NAME]
          --engine.github.repo.owner=          owner of the repository [$ENGINE_GITHUB_REPO_OWNER]
          --engine.github.repo.name=           name of the repository [$ENGINE_GITHUB_REPO_NAME]

    basic-auth:
          --engine.github.basic-auth.username= username for basic auth [$ENGINE_GITHUB_BASIC_AUTH_USERNAME]
          --engine.github.basic-auth.password= password for basic auth [$ENGINE_GITHUB_BASIC_AUTH_PASSWORD]

    gitlab:
          --engine.gitlab.token=               token to connect to the gitlab repository [$ENGINE_GITLAB_TOKEN]
          --engine.gitlab.base-url=            base url of the gitlab instance [$ENGINE_GITLAB_BASE_URL]
          --engine.gitlab.project-id=          project id of the repository [$ENGINE_GITLAB_PROJECT_ID]
          --engine.gitlab.timeout=             timeout for http requests (default: 5s) [$ENGINE_GITLAB_TIMEOUT]

    notify:
          --notify.stdout                      print release notes to stdout [$NOTIFY_STDOUT]
          --notify.stderr                      print release notes to stderr [$NOTIFY_STDERR]

    telegram:
          --notify.telegram.chat-id=           id of the chat, where the release notes will be sent [$NOTIFY_TELEGRAM_CHAT_ID]
          --notify.telegram.token=             bot token [$NOTIFY_TELEGRAM_TOKEN]
          --notify.telegram.web-page-preview   request telegram to preview for web links [$NOTIFY_TELEGRAM_WEB_PAGE_PREVIEW]
          --notify.telegram.timeout=           timeout for http requests (default: 5s) [$NOTIFY_TELEGRAM_TIMEOUT]

    github:
          --notify.github.timeout=             timeout for http requests (default: 5s) [$NOTIFY_GITHUB_TIMEOUT]
          --notify.github.release-name-tmpl=   template for release name [$NOTIFY_GITHUB_RELEASE_NAME_TMPL]
          --notify.github.tag=                 tag to specify release [$NOTIFY_GITHUB_TAG]
          --notify.github.extra=               extra parameters to pass to the notifier [$NOTIFY_GITHUB_EXTRA]

    repo:
          --notify.github.repo.full-name=      full name of the repository (owner/name) [$NOTIFY_GITHUB_REPO_FULL_NAME]
          --notify.github.repo.owner=          owner of the repository [$NOTIFY_GITHUB_REPO_OWNER]
          --notify.github.repo.name=           name of the repository [$NOTIFY_GITHUB_REPO_NAME]

    basic-auth:
          --notify.github.basic-auth.username= username for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_USERNAME]
          --notify.github.basic-auth.password= password for basic auth [$NOTIFY_GITHUB_BASIC_AUTH_PASSWORD]

    mattermost-hook:
          --notify.mattermost-hook.url=        url of the mattermost hook [$NOTIFY_MATTERMOST_HOOK_URL]
          --notify.mattermost-hook.timeout=    timeout for http requests (default: 5s) [$NOTIFY_MATTERMOST_HOOK_TIMEOUT]

    mattermost-bot:
          --notify.mattermost-bot.base-url=    base url for mattermost API [$NOTIFY_MATTERMOST_BOT_BASE_URL]
          --notify.mattermost-bot.token=       token of the mattermost bot [$NOTIFY_MATTERMOST_BOT_TOKEN]
          --notify.mattermost-bot.channel-id=  channel id of the mattermost bot [$NOTIFY_MATTERMOST_BOT_CHANNEL_ID]
          --notify.mattermost-bot.timeout=     timeout for http requests (default: 5s) [$NOTIFY_MATTERMOST_BOT_TIMEOUT]

    post:
          --notify.post.url=                   url to send the release notes [$NOTIFY_POST_URL]
          --notify.post.timeout=               timeout for http requests (default: 5s) [$NOTIFY_POST_TIMEOUT]

    task:
          --task.type=[|jira]                  type of the task tracker [$TASK_TYPE]

    jira:
          --task.jira.base-url=                url of the jira instance [$TASK_JIRA_BASE_URL]
          --task.jira.token=                   token to connect to the jira instance [$TASK_JIRA_TOKEN]
          --task.jira.timeout=                 timeout for http requests (default: 5s) [$TASK_JIRA_TIMEOUT]

    enricher:
          --task.jira.enricher.load-watchers   load watchers for the issue [$TASK_JIRA_ENRICHER_LOAD_WATCHERS]
```

</details>

**Note**: `from` and `to` options of `changelog` command accept expressions, which must be written in gotemplate manner.
`from` expression can relate on the value of `to`, `to` is evaluated first.

Example (from .env file): `TO='{{ last_commit "develop" }}'`

## Preview data file structure

<details>

<summary>Click to expand</summary>

| Field                            | Description                                                                      |
|----------------------------------|----------------------------------------------------------------------------------|
| from                             | Commit ref to start release notes from                                           |
| to                               | Commit ref to end release notes at                                               |
| extras                           | Extra variables to use in the template                                           |
| pull_requests.number             | Pull request number                                                              |
| pull_requests.title              | Pull request title                                                               |
| pull_requests.body               | Pull request body                                                                |
| pull_requests.author.username    | Pull request's author's username                                                 |
| pull_requests.author.email       | Pull request's author's email                                                    |
| pull_requests.labels             | List of pull request's labels                                                    |
| pull_requests.closed_at          | Date of the pull request's closing                                               |
| pull_requests.source_branch      | Pull request's source branch                                                     |
| pull_requests.target_branch      | Pull request's target branch                                                     |
| pull_requests.url                | Pull request's url                                                               |
| pull_requests.received_by_shas   | List of commit SHAs by which pull request was retrieved (for debugging purposes) |
| pull_requests.assignees.username | Assignee's username                                                              |
| pull_requests.assignees.email    | Assignee's email                                                                 |
| tasks.id                         | Task ID                                                                          |
| tasks.parent_id                  | Task's parent ID                                                                 |
| tasks.url                        | Task's URL                                                                       |
| tasks.name                       | Task's name                                                                      |
| tasks.body                       | Task's body                                                                      |
| tasks.closed_at                  | Task's closing date                                                              |
| tasks.author.username            | Task's author's username                                                         |
| tasks.author.email               | Task's author's email                                                            |
| tasks.assignee.username          | Task's assignee's username                                                       |
| tasks.assignee.email             | Task's assignee's email                                                          |
| tasks.type                       | Task's type                                                                      |
| tasks.type_raw                   | Task's raw type                                                                  |
| tasks.flagged                    | Task's flagged status                                                            |
| tasks.watchers.username          | List of task's watchers                                                          |
| tasks.watchers.email             | List of task's watchers' emails                                                  |
| tasks.watches_count              | Count of task's watchers                                                         |
| commits.sha                      | Commit SHA                                                                       |
| commits.parent_shas              | List of commit's parent SHAs                                                     |
| commits.message                  | Commit message                                                                   |
| commits.committed_at             | Commit's commit date                                                             |
| commits.authored_at              | Commit's authored date                                                           |
| commits.url                      | Commit's URL                                                                     |
| commits.author.username          | Commit's author's username                                                       |
| commits.author.email             | Commit's author's email                                                          |
| commits.committer.username       | Commit's committer's username                                                    |
| commits.committer.email          | Commit's committer's email                                                       |

</details>

See [example](_example/simple-prs/preview_data.yaml) for details.

## Evaluator functions

The list of available to use functions consists of [sprig's](http://masterminds.github.io/sprig/) functions list, and a several custom functions, including:

<details>

<summary>Click to expand</summary>

| Function name                                                                                                           | Description                                                                                                                                      |
|-------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| `next(elem string, elems []string) string`                                                                              | returns next element in the list                                                                                                                 |
| `previous(elem string, elems []string) string`                                                                          | returns previous element in the list                                                                                                             |
| `filter(rx string, elems []string) []string`                                                                            | filters list of strings by regular expression                                                                                                    |
| `stringsFromAnys([]interface{}) []string`                                                                               | casts list of `any` to list of strings                                                                                                           |
| **_Constants_**                                                                                                         |                                                                                                                                                  |
| `semver() string`                                                                                                       | returns semver regular expression  (`^v?(\d+)\.(\d+)\.(\d+)$`)                                                                                   |
| **_Addons_**                                                                                                            |                                                                                                                                                  |
| **git**                                                                                                                 |                                                                                                                                                  |
| `prTitles(prs []git.PullRequest) []string`                                                                              | returns the list of titles of provided pull requests                                                                                             |
| `headed(vals []string) []string`                                                                                        | adds "HEAD" to the head of the list                                                                                                              |
| `lastCommit(branch string) (string, error)`                                                                             | returns the SHA of the last commit of the branch                                                                                                 |
| `tags() ([]string, error)`                                                                                              | returns the list of all tags in repository                                                                                                       |
| `previousTag(commitAlias string, tags []string) (string, error)`                                                        | returns the previous (closest) tag of the provided commit in the provided list of tags                                                           |
| **task**                                                                                                                |                                                                                                                                                  |
| `getTicket(id string) (task.Ticket, error)`                                                                             | returns ticket by its ID                                                                                                                         |
| `listTickets(ids []string, loadParents bool) ([]task.Ticket, error)`                                                    | lists tickets by their IDs, with parents attached, if `loadParents` is set to true                                                               |
| **release-notes**                                                                                                       |                                                                                                                                                  |
| `buildTicketsTree(tickets []task.Ticket) (roots []*TicketNode, err error)`                                              | builds tree out of provided tickets                                                                                                              |
| `loadTicketsTree(ticketIDRx string, loadParents bool, prs []git.PullRequest, commits []git.Commit) (LoadedTree, error)` | loads tickets tree from the provided pull requests, ticket IDs are matched from pull request titles by the provided regexp                       |
| `listTaskUsers(obj any, args ...string) (string, error)`                                                                | lists users from the provided task ticket, any in first argument to match embedded structs                                                       |
| `listPRs(prs []git.PullRequest) string`                                                                                 | returns a comma-separated list of markdown-formatted links to PRs, example: `[Title1](URL1), [Title2](URL2)`                                     |
| `listCommits(commits []git.Commit) string`                                                                              | returns a comma-separated list of markdown-formatted links to commits, example: `[short-SHA1](URL1), [short-SHA2](URL2)`                         |
| `mdTaskLink(obj any) (string, error)`                                                                                   | returns markdown-formatted link to the task ticket, example: `[Title](URL)`                                                                      |
| `brackets(s string, square ...bool) string`                                                                             | returns string wrapped in brackets, if `square` is set to true, square brackets will be used. If the string is empty, it returns an empty string |
| `log(msg string, args ...interface{}) string`                                                                           | logs the debug message with provided arguments to stderr. Always returns an empty string                                                         |

</details>

- `from` and `to` flags may use `git` addon functions.
- release notes builder template may use all functions above.

### Types

<details>

<summary>Click to expand</summary>

```go
// LoadedTree is a tree of tickets with their children and PRs.
type LoadedTree struct {
	Roots      []*TicketNode
	Unattached []git.PullRequest
}

// TicketNode is a representation of a ticket with its children.
type TicketNode struct {
	task.Ticket
	Children []*TicketNode
	PRs      []git.PullRequest
}

// PullRequest represents a pull/merge request from the
// remote repository.
type PullRequest struct {
	Number         int       `yaml:"number"`
	Title          string    `yaml:"title"`
	Body           string    `yaml:"body"`
	Author         User      `yaml:"author"`
	Labels         []string  `yaml:"labels"`
	ClosedAt       time.Time `yaml:"closed_at"`
	SourceBranch   string    `yaml:"source_branch"`
	TargetBranch   string    `yaml:"target_branch"`
	URL            string    `yaml:"url"`
	ReceivedBySHAs []string  `yaml:"received_by_shas"`
	Assignees      []User    `yaml:"assignees"`
}

// User holds user data.
type User struct {
	Username string `yaml:"username"`
	Email    string `yaml:"email"`
}

// Commit represents a repository commit.
type Commit struct {
	SHA         string
	ParentSHAs  []string
	Message     string
	CommittedAt time.Time
	AuthoredAt  time.Time
}

// CommitsComparison is the result of comparing two commits.
type CommitsComparison struct {
	Commits      []Commit
	TotalCommits int
}

// Tag represents a repository tag.
type Tag struct {
	Name   string
	Commit Commit
}

// Type specifies the type of the task.
type Type string

const (
    // TypeEpic is an epic task type.
    TypeEpic Type = "epic"
    // TypeTask is a simple task type.
    TypeTask Type = "task"
    // TypeSubtask is a sub-task type.
    TypeSubtask Type = "subtask"
)

// Ticket represents a single task in task tracker.
type Ticket struct {
    ID       string
    ParentID string
    
    URL      string
    Name     string
    Body     string
    ClosedAt time.Time
    Author   User
    Assignee User
    Type     Type
    TypeRaw  string // save raw type in case if user wants to distinguish different raw values
    Flagged  bool
}

// User represents a task tracker user.
type User struct {
    Username string
    Email    string
}


// LoadedTree is a tree of tickets with their children and PRs.
type LoadedTree struct {
    Roots      []*TicketNode
    Unattached []git.PullRequest
}

// TicketNode is a representation of a ticket with its children.
type TicketNode struct {
    task.Ticket
    Children []*TicketNode
    PRs      []git.PullRequest
}
```

</details>

## Release notes builder configuration
| Name                      | Description                                                                                                                                             |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| categories                | Categories of pull requests                                                                                                                             |
| categories.title          | Title, which will be provided to the release notes template                                                                                             |
| categories.labels         | An array of labels, to match pull request labels against. If any PR label matches any category label, the pull request will show up under this category |
| categories.branch         | A regular expression to match source branch name to the corresponding category.                                                                         |
| categories.commit_message | A regular expression to match commit message to the corresponding category.                                                                             |
| sort_field                | Field, by which pull requests must be sorted, in format +&#124;-field currently supported fields: `number`, `author`, `title`, `closed`                 |
| template                  | Template for a changelog in golang's text template language                                                                                             |
| unused_title              | If set, the unused category will be built under this title at the end of the changelog                                                                  |
| ignore_labels             | An array of labels, to match pull request labels against. If PR contains any of the defined ignore labels - this PR won't be provided to the template   |
| ignore_branch             | A regular expression to match pull request branches, that won't appear in the changelog                                                                 |

See [example](_example/simple-prs/config.yaml) for details.

## Template variables for release notes builder

**NOTE:** commits will be provided in categories only in case if it wasn't matched to any pull request.

| Name                                | Description                                                    | Example                                         |
|-------------------------------------|----------------------------------------------------------------|-------------------------------------------------|
| {{.From}}                           | From commit SHA / tag                                          | v0.1.0                                          |
| {{.To}}                             | To commit SHA / tag                                            | v0.2.0                                          |
| {{.Date}}                           | Date, when the changelog was built                             | Jan 02, 2006 15:04:05 UTC                       |
| {{.Extras}}                         | Map of extra variables, provided by the user in envs           | map[foo:bar]                                    |
| {{.Total}}                          | Total number of pull requests                                  | 10                                              |
| {{.Categories.Title}}               | Title of the category from the config                          | Features                                        |
| {{.Categories.PRs.Number}}          | Number of the pull request                                     | 642                                             |
| {{.Categories.PRs.Title}}           | Title of the pull request                                      | Some awesome feature added                      |
| {{.Categories.PRs.Author}}          | Username of the author of pull request                         | Semior001                                       |
| {{.Categories.PRs.URL}}             | URL to the pull request                                        | `https://github.com/Semior001/releaseit/pull/6` |
| {{.Categories.PRs.SourceBranch}}    | Source branch name, from which the pull request was created    | feature/awesome-feature                         |
| {{.Categories.PRs.TargetBranch}}    | Target branch name, to which the pull request was created      | develop                                         |
| {{.Categories.PRs.ClosedAt}}        | Timestamp, when the pull request was closed (might be empty)   | Jan 02, 2006 15:04:05 UTC                       |
| {{.Categories.PRs.ReceivedBySHAs}}  | List of commit SHAs, by which releaseit received pull requests | [a1b2c3d4e5f6, 1a2b3c4d5e6f]                    |
| {{.Categories.PRs.Assignees}}       | List of assignees of the pull request                          | [Semior001, Semior002]                          |
| {{.Categories.Commits.SHA}}         | SHA of the commit                                              | a1b2c3d4e5f6                                    |
| {{.Categories.Commits.ParentSHAs}}  | List of parent commit SHAs                                     | [a1b2c3d4e5f6, 1a2b3c4d5e6f]                    |
| {{.Categories.Commits.Message}}     | Message of the commit                                          | some feature merged                             |
| {{.Categories.Commits.CommittedAt}} | Timestamp, when the commit was committed                       | Jan 02, 2006 15:04:05 UTC                       |
| {{.Categories.Commits.AuthoredAt}}  | Timestamp, when the commit was authored                        | Jan 02, 2006 15:04:05 UTC                       |
| {{.Categories.Commits.URL}}         | URL to the commit                                              | `                                               |
| {{.Categories.Commits.Author}}      | Username of the author of the commit                           | Semior001                                       |
| {{.Categories.Commits.Committer}}   | Username of the committer of the commit                        | Semior001                                       |

For functions available to use see the [list of evaluator functions](#evaluator-functions).

## (Github) Template variables for release title

| Name                       | Description                     | Example             |
|----------------------------|---------------------------------|---------------------|
| {{.Tag.Name}}              | Tag name                        | v1.0.0              |
| {{.Tag.Message}}           | Tag message                     | Version v1.0.0      |
| {{.Tag.Author}}            | Tag author                      | Semior001           |
| {{.Tag.Date}}              | Tag date                        | Jan 02, 2006 15:04  |
| {{.Commit.SHA}}            | Commit SHA                      | a1b2c3d4e5f6        |
| {{.Commit.Message}}        | Commit message                  | some feature merged |
| {{.Commit.Author.Name}}    | Commit author name              | Semior001           |
| {{.Commit.Author.Date}}    | Date, when commit was authored  | Jan 02, 2006 15:04  |
| {{.Commit.Committer.Name}} | Commit committer name           | Semior001           |
| {{.Commit.Committer.Date}} | Date, when commit was committed | Jan 02, 2006 15:04  |
| {{.Extras}}                | Map of extra variables          | map[foo:bar]        |

For functions available to use see the [list of evaluator functions](#evaluator-functions).
