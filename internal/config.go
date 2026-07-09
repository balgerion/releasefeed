package internal

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Apprise struct {
	URL  string   `yaml:"url"`
	Tags []string `yaml:"tags"`
}

type Defaults struct {
	FetchInterval   string   `yaml:"fetch_interval"`
	AlertKeywords   []string `yaml:"alert_keywords"`
	ExcludeSections []string `yaml:"exclude_sections"`
	Image           string   `yaml:"image"`
}

type Feed struct {
	Name            string           `yaml:"name"`
	URL             string           `yaml:"url"`
	FetchInterval   string           `yaml:"fetch_interval"`
	ExcludeSections []string         `yaml:"exclude_sections"`
	RegexRemove     []string         `yaml:"regex_remove"`
	Image           string           `yaml:"image"`
	AlertKeywords   []string         `yaml:"alert_keywords"`
	AppriseTags     []string         `yaml:"apprise_tags"`
	regexRemove     []*regexp.Regexp `yaml:"-"`
}

type Server struct {
	Port    string `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type Config struct {
	Server   Server   `yaml:"server"`
	Apprise  Apprise  `yaml:"apprise"`
	Defaults Defaults `yaml:"defaults"`
	Feeds    []Feed   `yaml:"feeds"`
}

func merge(a, b []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, s := range append(a, b...) {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
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
		if cfg.Feeds[i].Image == "" {
			cfg.Feeds[i].Image = cfg.Defaults.Image
		}
		cfg.Feeds[i].AlertKeywords = merge(cfg.Defaults.AlertKeywords, cfg.Feeds[i].AlertKeywords)
		cfg.Feeds[i].ExcludeSections = merge(cfg.Defaults.ExcludeSections, cfg.Feeds[i].ExcludeSections)
		if len(cfg.Feeds[i].AppriseTags) == 0 {
			cfg.Feeds[i].AppriseTags = cfg.Apprise.Tags
		}
		for _, pattern := range cfg.Feeds[i].RegexRemove {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("feed %s: invalid regex_remove %q: %w", cfg.Feeds[i].Name, pattern, err)
			}
			cfg.Feeds[i].regexRemove = append(cfg.Feeds[i].regexRemove, re)
		}
	}
	return &cfg, nil
}