package sentry

import (
	"log"
	"net/url"

	"github.com/canva/terraform-provider-sentry/sentryclient"
)

// Config is the configuration structure used to instantiate the Sentry
// provider.
type Config struct {
	Token   string
	BaseURL string
}

// Client to connect to Sentry.
func (c *Config) Client() (interface{}, error) {
	var baseURL *url.URL
	var err error

	if c.BaseURL != "" {
		baseURL, err = url.Parse(c.BaseURL)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[INFO] Instantiating Sentry client...")
	cl := sentryclient.NewClient(nil, baseURL, c.Token)

	return cl, nil
}
