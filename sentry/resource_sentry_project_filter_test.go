package sentry

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jianyuan/go-sentry/v2/sentry"
)

func TestAccSentryProjectFilter_basic(t *testing.T) {
	teamName := acctest.RandomWithPrefix("tf-team")
	projectName := acctest.RandomWithPrefix("tf-project")
	var filterConfig sentry.FilterConfig

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSentryProjectFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSentryProjectFilterConfig(teamName, projectName),
				Check:  testFilterConfig("sentry_filter.test_filter", &filterConfig, true, []string{"ie_pre_9", "ie10"}),
			},
		},
	})
}

func testAccCheckSentryProjectFilterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*sentry.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sentry_filter" {
			continue
		}

		ctx := context.Background()
		filterConfig, _, err := client.ProjectFilter.GetFilterConfig(ctx, testOrganization, rs.Primary.Attributes["project"])
		// We should not be able to reach https://[API]/[PROJECT]/filters since it should be deleted at this point.
		// TODO: Don't error out in `go-sentry.ProjectFilter.GetFilterConfig` if the project or rule does not exist.
		if strings.Contains(err.Error(), "The requested resource does not exist") {
			return nil
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Received a sentry_filter, but it should have been deleted. Filter: %v", filterConfig)
	}

	return nil
}

func testFilterConfig(n string, filterConfig *sentry.FilterConfig, browserExtension bool, legacyBrowsers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No project ID is set")
		}

		ctx := context.Background()
		client := testAccProvider.Meta().(*sentry.Client)
		filterConfig, _, err := client.ProjectFilter.GetFilterConfig(ctx, testOrganization, rs.Primary.Attributes["project"])
		if err != nil {
			return err
		}
		if filterConfig.BrowserExtension != browserExtension {
			return fmt.Errorf("got browser_extension %t; want %t", filterConfig.BrowserExtension, browserExtension)
		}

		if !cmp.Equal(filterConfig.LegacyBrowsers, legacyBrowsers, cmp.Transformer("sort", func(in []string) []string {
			sort.Strings(in)
			return in
		})) {
			return fmt.Errorf("got legacy_browser %v; want %v", filterConfig.LegacyBrowsers, legacyBrowsers)
		}

		return nil
	}
}

func testAccSentryProjectFilterConfig(teamName string, projectName string) string {
	return testAccSentryProjectConfig(teamName, projectName) + fmt.Sprintf(`
resource "sentry_filter" "test_filter" {
	organization = "%s"
	project = sentry_project.test.id
	browser_extension = true
  legacy_browsers = ["ie_pre_9","ie10"]
}
`, testOrganization)
}
