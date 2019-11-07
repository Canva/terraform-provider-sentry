package main

import (
	"github.com/fa93hws/terraform-provider-sentry/sentry"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sentry.Provider,
	})
}
