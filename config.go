package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type DateRange struct {
	StartDate string `yaml:"start_date"` // YYYYMMDD format
	EndDate   string `yaml:"end_date"`   // YYYYMMDD format
	Service   string `yaml:"service"`    // Target service URL
}

type ParsedDateRange struct {
	StartDate  time.Time
	EndDate    time.Time
	ServiceURL *url.URL
}

type Config struct {
	Port         int         `yaml:"port"`
	ReadTimeout  string      `yaml:"read_timeout"`
	WriteTimeout string      `yaml:"write_timeout"`
	IdleTimeout  string      `yaml:"idle_timeout"`
	DateRanges   []DateRange `yaml:"date_ranges"`
}

type ParsedConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	DateRanges   []ParsedDateRange
}

func LoadConfig(configPath string) (*ParsedConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return parseConfig(&config)
}

func parseConfig(config *Config) (*ParsedConfig, error) {
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", config.Port)
	}

	if len(config.DateRanges) == 0 {
		return nil, fmt.Errorf("no date ranges configured")
	}

	readTimeout, err := time.ParseDuration(config.ReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid read_timeout: %w", err)
	}

	writeTimeout, err := time.ParseDuration(config.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid write_timeout: %w", err)
	}

	idleTimeout, err := time.ParseDuration(config.IdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid idle_timeout: %w", err)
	}

	parsedRanges := make([]ParsedDateRange, 0, len(config.DateRanges))
	for i, dr := range config.DateRanges {
		startDate, err := time.Parse("20060102", dr.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format in range %d: %s", i, dr.StartDate)
		}

		endDate, err := time.Parse("20060102", dr.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format in range %d: %s", i, dr.EndDate)
		}

		serviceURL, err := url.Parse(dr.Service)
		if err != nil {
			return nil, fmt.Errorf("invalid service URL in range %d: %w", i, err)
		}

		parsedRanges = append(parsedRanges, ParsedDateRange{
			StartDate:  startDate,
			EndDate:    endDate,
			ServiceURL: serviceURL,
		})
	}

	return &ParsedConfig{
		Port:         config.Port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		DateRanges:   parsedRanges,
	}, nil
}
