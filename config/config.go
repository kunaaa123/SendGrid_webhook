package config

type Config struct {
	LarkWebhookURL string
	ServerPort     string
}

func NewConfig() *Config {
	return &Config{
		LarkWebhookURL: "https://open.larksuite.com/open-apis/bot/v2/hook/88fccfea-8fad-47d9-99a9-44d214785fff",
		ServerPort:     ":8080",
	}
}
