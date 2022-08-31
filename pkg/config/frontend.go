package config

type FrontendConfig struct {
	TagName     string `mapstructure:"tag_name" json:"tag_name"`
	Commit      string `mapstructure:"target_commitish" json:"target_commitish"`
	Name        string `mapstructure:"name" json:"name"`
	Draft       bool   `mapstructure:"draft" json:"draft"`
	Prerelease  bool   `mapstructure:"prerelease" json:"prerelease"`
	CreatedAt   string `mapstructure:"created_at" json:"created_at"`
	PublishedAt string `mapstructure:"published_at" json:"published_at"`
}

func GetFrontend() FrontendConfig {
	return store.Frontend
}

func SetFrontend(f FrontendConfig) error {
	mu.Lock()
	defer mu.Unlock()
	store.Frontend = f
	return saveConfig()
}
