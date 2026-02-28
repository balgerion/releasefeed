package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Apprise struct {
	URL  string   `yaml:"url"`
	Tags []string `yaml:"tags"`
}

type Defaults struct {
	FetchInterval string   `yaml:"fetch_interval"`
	AlertKeywords []string `yaml:"alert_keywords"`
}

type Sections struct {
	Mode  string   `yaml:"mode"`
	Names []string `yaml:"names"`
}

type Feed struct {
	Name          string   `yaml:"name"`
	URL           string   `yaml:"url"`
	FetchInterval string   `yaml:"fetch_interval"`
	Sections      Sections `yaml:"sections"`
	RegexRemove   []string `yaml:"regex_remove"`
	Raw           bool     `yaml:"raw"`
	AlertKeywords []string `yaml:"alert_keywords"`
	AppriseTags   []string `yaml:"apprise_tags"`
}

type Server struct {
	Port string `yaml:"port"`
}

type Config struct {
	Server   Server   `yaml:"server"`
	Apprise  Apprise  `yaml:"apprise"`
	Defaults Defaults `yaml:"defaults"`
	Feeds    []Feed   `yaml:"feeds"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Defaults.FetchInterval == "" {
		cfg.Defaults.FetchInterval = "1h"
	}
	for i := range cfg.Feeds {
		if cfg.Feeds[i].FetchInterval == "" {
			cfg.Feeds[i].FetchInterval = cfg.Defaults.FetchInterval
		}
		if len(cfg.Feeds[i].AlertKeywords) == 0 {
			cfg.Feeds[i].AlertKeywords = cfg.Defaults.AlertKeywords
		}
		if len(cfg.Feeds[i].AppriseTags) == 0 {
			cfg.Feeds[i].AppriseTags = cfg.Apprise.Tags
		}
	}
	return &cfg, nil
}
