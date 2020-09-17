// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package config manages configuration files for the twig tool.
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/creachadair/atomicfile"
	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/auth"
	yaml "gopkg.in/yaml.v3"
)

// Config represents the stored configuration data for the twig tools.
type Config struct {
	// Required fields. See: https://developer.twitter.com/en/portal/dashboard
	APIKey      string `yaml:"api_key"`
	APISecret   string `yaml:"api_secret"`
	Token       string `yaml:"access_token"`
	Secret      string `yaml:"access_secret"`
	BearerToken string `yaml:"bearer_token,omitempty"`

	Users []*User `yaml:"users,omitempty"`

	// Non-persistent fields.
	filePath string
	AuthUser string                             `yaml:"-"`
	Log      func(tag jhttp.LogTag, msg string) `yaml:"-"`
	LogMask  jhttp.LogTag                       `yaml:"-"`
}

// User carries an access token for an individual user.
type User struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Secret   string `yaml:"secret"`
}

// NewClient returns a new Twitter client from selected settings.
func (c *Config) NewClient() (*twitter.Client, error) {
	if c.AuthUser != "" {
		return c.NewUserClient(c.AuthUser)
	}
	return c.NewBearerClient()
}

// NewBearerClient returns a new Twitter client with a bearer token.
func (c *Config) NewBearerClient() (*twitter.Client, error) {
	if c.BearerToken == "" {
		return nil, errors.New("no bearer token is available")
	}
	return twitter.NewClient(&jhttp.Client{
		Authorize: jhttp.BearerTokenAuthorizer(c.BearerToken),
		Log:       c.Log,
		LogMask:   c.LogMask,
	}), nil
}

// NewUserClient returns a new Twitter client with an access token for the
// specified username.
func (c *Config) NewUserClient(user string) (*twitter.Client, error) {
	u := c.FindUsername(user)
	if u == nil {
		return nil, fmt.Errorf("no access token foundfor user %q", user)
	}
	cfg := auth.Config{APIKey: c.APIKey, APISecret: c.APISecret}
	return twitter.NewClient(&jhttp.Client{
		Authorize: cfg.Authorizer(u.Token, u.Secret),
		Log:       c.Log,
	}), nil
}

// FindUsername returns the access token for the given username, or nil.
func (c *Config) FindUsername(name string) *User {
	needle := strings.ToLower(name)
	for _, u := range c.Users {
		if strings.ToLower(u.Username) == needle {
			return u
		}
	}
	return nil
}

// Save writes the current state of c back to its original file.
func (c *Config) Save() error {
	if c.filePath == "" {
		return errors.New("unknown file path")
	}
	return Save(c, c.filePath)
}

// Load reads in the contents of a config file from path.
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decoding config data: %w", err)
	}
	cfg.filePath = path
	return &cfg, nil
}

// Save writes cfg to path.
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return atomicfile.WriteData(path, data, 0600)
}
