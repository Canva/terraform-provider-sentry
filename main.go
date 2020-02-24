package main

import (
	"github.com/canva/terraform-provider-sentry/sentry"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sentry.Provider,
	})
}
