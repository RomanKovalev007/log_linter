package loglint

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the linter configuration.
type Config struct {
	Rules    RulesConfig `yaml:"rules"`
	Keywords []string    `yaml:"sensitive_keywords"`
}

// RulesConfig controls which rules are enabled.
type RulesConfig struct {
	Lowercase    *bool `yaml:"lowercase"`
	EnglishOnly  *bool `yaml:"english_only"`
	NoSpecial    *bool `yaml:"no_special_chars"`
	SensitiveData *bool `yaml:"sensitive_data"`
}

func defaultConfig() Config {
	t := true
	return Config{
		Rules: RulesConfig{
			Lowercase:     &t,
			EnglishOnly:   &t,
			NoSpecial:     &t,
			SensitiveData: &t,
		},
		Keywords: defaultSensitiveKeywords,
	}
}

var defaultSensitiveKeywords = []string{
	"password", "pwd",
	"secret", "token",
	"api_key", "apikey",
	"private_key", "privatekey",
	"access_key", "accesskey",
	"credential", "bearer",
	"session_id",
}

func (c Config) isLowercaseEnabled() bool {
	return c.Rules.Lowercase == nil || *c.Rules.Lowercase
}

func (c Config) isEnglishOnlyEnabled() bool {
	return c.Rules.EnglishOnly == nil || *c.Rules.EnglishOnly
}

func (c Config) isNoSpecialEnabled() bool {
	return c.Rules.NoSpecial == nil || *c.Rules.NoSpecial
}

func (c Config) isSensitiveDataEnabled() bool {
	return c.Rules.SensitiveData == nil || *c.Rules.SensitiveData
}

func (c Config) sensitiveKeywords() []string {
	if len(c.Keywords) > 0 {
		return c.Keywords
	}
	return defaultSensitiveKeywords
}

func loadConfig(path string) (Config, error) {
	if path == "" {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
