package sentry

import (
	"fmt"
	"testing"

	"github.com/canva/terraform-provider-sentry/sentryclient"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccSentryFilterUpdate(t *testing.T) {
	var filterConfig sentryclient.FilterConfig
	testAccSentryFilterUpdateConfig := fmt.Sprintf(`
	resource "sentry_filter" "test_filter" {
		organization = "%s"
		project = "%s"
		browser_extension = true
		legacy_browsers = ["ie_pre_9","ie10"]
	}
	`, testOrganization, projectSlug)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSentryFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSentryFilterUpdateConfig,
				Check:  testFilterConfig(&filterConfig, true, []string{"ie_pre_9", "ie10"}),
			},
		},
	})
}

func contains(array []string, element string) bool {
	for _, a := range array {
		if a == element {
			return true
		}
	}
	return false
}

func testFilterConfig(filterConfig *sentryclient.FilterConfig, browserExtension bool, legacyBrowsers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*sentryclient.Client)
		filterConfig, _, err := client.ProjectFilter.Get(testOrganization, projectSlug)
		if err != nil {
			return err
		}
		if filterConfig.BrowserExtension != browserExtension {
			return fmt.Errorf("got browser_extension %t; want %t", filterConfig.BrowserExtension, browserExtension)
		}

		for _, browser := range legacyBrowsers {
			if !contains(filterConfig.LegacyBrowsers, browser) {
				return fmt.Errorf("got legacy_browser %v; want %v", filterConfig.LegacyBrowsers, legacyBrowsers)
			}
		}

		return nil
	}
}

func testAccCheckSentryFilterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*sentryclient.Client)
	filterConfig, _, err := client.ProjectFilter.Get(testOrganization, projectSlug)
	if err != nil {
		return err
	}
	if filterConfig.BrowserExtension != false {
		return fmt.Errorf("got browser_extension %t; want false", filterConfig.BrowserExtension)
	}
	if len(filterConfig.LegacyBrowsers) != 0 {
		return fmt.Errorf("got legacy_browser %v; want []", filterConfig.LegacyBrowsers)
	}
	return nil
}
