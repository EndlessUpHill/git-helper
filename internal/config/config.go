package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	GithubToken string `mapstructure:"github_token"`
	DefaultOrg  string `mapstructure:"default_org"`
	Debug       bool   `mapstructure:"debug"`
}

func LoadConfig(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".githelper")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GITHELPER")

	var config Config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
} 