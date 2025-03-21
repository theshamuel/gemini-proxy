package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type Config struct {
	FileName string
	sync.Mutex
	File *File
}

type File struct {
	GeminiAPIKey  string `yaml:"gemini-api-key"`
	DelayRequests int    `yaml:"delay-requests"`
	TLS           struct {
		Enabled        bool   `yaml:"enabled,omitempty"`
		CertPath       string `yaml:"cert-path,omitempty"`
		PrivateKeyPath string `yaml:"private-key-path,omitempty"`
	} `yaml:"tls,omitempty"`
	Debug bool `yaml:"debug,omitempty"`
}

type CommonOpts struct {
	GeminiAPIKey  string `long:"geminiAPIKey" env:"GEMINI_API_KEY" description:"the key to access Gemini API"`
	DelayRequests int    `long:"delayRequests" env:"DELAY_REQUESTS" default:"0" description:"the delay between requests if 0 no delay"`
	TLS           TLS    `group:"tls" namespace:"tls" env-namespace:"TLS"`
	Debug         bool   `long:"debug" env:"DEBUG" description:"debug mode"`
}

type TLS struct {
	Enabled        bool   `long:"enabled" env:"ENABLED" description:"Enable TLS support."`
	CertPath       string `long:"cert" env:"CERT" default:"domain.crt" description:"Set certificate path for TLS support"`
	PrivateKeyPath string `long:"private-key" env:"PRIVATE_KEY" default:"default.key" description:"Set private key path for TLS support"`
}

func (s *Config) GetCommon() (*CommonOpts, error) {
	s.Lock()
	defer s.Unlock()
	f, err := os.Open(s.FileName)
	if err != nil {
		return nil, fmt.Errorf("can't open %s: %w", s.FileName, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("can't close config file %s: %v\n", s.FileName, err)
		}
	}(f)
	if err = yaml.NewDecoder(f).Decode(&s.File); err != nil {
		return nil, fmt.Errorf("can't parse %s: %w", s.FileName, err)
	}

	return &CommonOpts{
		GeminiAPIKey:  s.File.GeminiAPIKey,
		DelayRequests: s.File.DelayRequests,
		TLS: TLS{
			Enabled:        s.File.TLS.Enabled,
			CertPath:       s.File.TLS.CertPath,
			PrivateKeyPath: s.File.TLS.PrivateKeyPath,
		},
		Debug: s.File.Debug,
	}, nil
}
