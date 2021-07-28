package cmd

// GithubGroup defines parameters to connect to the github repository
type GithubGroup struct {
	Repo struct {
		Owner string `long:"owner" env:"OWNER" description:"owner of the repository" required:"true"`
		Name  string `long:"name" env:"NAME" description:"name of the repository" required:"true"`
	} `group:"repo" namespace:"repo" env-namespace:"REPO"`
	BasicAuth BasicAuthGroup `group:"basic_auth" namespace:"basic_auth" env-namespace:"BASIC_AUTH"`
}

// BasicAuthGroup defines parameters for basic authentication.
type BasicAuthGroup struct {
	Username string `long:"username" env:"USERNAME" description:"username for basic auth"`
	Password string `long:"password" env:"PASSWORD" description:"password for basic auth"`
}

// TelegramGroup defines parameters for telegram notifier.
type TelegramGroup struct {
	ChatID                string `long:"chat_id" env:"CHAT_ID" description:"id of the chat, where the release notes will be sent"`
	Token                 string `long:"token" env:"TOKEN" description:"bot token"`
	DisableWebPagePreview bool   `long:"disable_web_page_preview" env:"DISABLE_WEB_PAGE_PREVIEW" description:"request telegram to disable preview for web links"`
}
