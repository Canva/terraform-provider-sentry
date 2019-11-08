package sentry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/canva/terraform-provider-sentry/sentryclient"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccSentryKey_basic(t *testing.T) {
	var key sentryclient.ProjectKey

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSentryKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSentryKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSentryKeyExists("sentry_key.test_key", &key),
					resource.TestCheckResourceAttr("sentry_key.test_key", "name", "Test key"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "public"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "secret"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_secret"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_public"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_csp"),
				),
			},
			{
				Config: testAccSentryKeyUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSentryKeyExists("sentry_key.test_key", &key),
					resource.TestCheckResourceAttr("sentry_key.test_key", "name", "Test key changed"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "public"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "secret"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_secret"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_public"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_csp"),
				),
			},
			{
				Config: testAccSentryRateLimitUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSentryKeyExists("sentry_key.test_key", &key),
					// This only work if the account has enterprise plan enabled
					testAccCheckSentryKeyAttributes(&key, &testAccSentryKeyExpectedAttributes{
						RateLimit: &sentryclient.ProjectKeyRateLimit{
							Count:  2000,
							Window: 300,
						},
					}),
					resource.TestCheckResourceAttr("sentry_key.test_key", "name", "Test key changed"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "public"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "secret"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_secret"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_public"),
					resource.TestCheckResourceAttrSet("sentry_key.test_key", "dsn_csp"),
				),
			},
		},
	})
}

func testAccCheckSentryKeyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*sentryclient.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sentry_key" {
			continue
		}

		keys, resp, err := client.ProjectKeys.List(
			rs.Primary.Attributes["organization"],
			rs.Primary.Attributes["project"],
		)
		if err == nil {
			for _, key := range keys {
				if key.ID == rs.Primary.ID {
					return errors.New("Key still exists")
				}
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckSentryKeyExists(n string, projectKey *sentryclient.ProjectKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No key ID is set")
		}

		client := testAccProvider.Meta().(*sentryclient.Client)
		keys, _, err := client.ProjectKeys.List(
			rs.Primary.Attributes["organization"],
			rs.Primary.Attributes["project"],
		)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if key.ID == rs.Primary.ID {
				*projectKey = key
				break
			}
		}
		return nil
	}
}

type testAccSentryKeyExpectedAttributes struct {
	RateLimit *sentryclient.ProjectKeyRateLimit
}

func checkRateLimit(get *sentryclient.ProjectKeyRateLimit, want *sentryclient.ProjectKeyRateLimit) error {
	if get == nil && want == nil {
		return nil
	}
	if get == nil && want != nil {
		return errors.New("got nil rate limit but want non-nil")
	}
	if get != nil && want == nil {
		return errors.New("got non-nil rate limit but want nil")
	}
	if get.Window != want.Window {
		return fmt.Errorf("got RateLimit.window %v; want %v", get.Window, want.Window)
	}
	if get.Count != want.Count {
		return fmt.Errorf("got RateLimit.window %v; want %v", get.Count, want.Count)
	}
	return nil
}

func testAccCheckSentryKeyAttributes(key *sentryclient.ProjectKey, want *testAccSentryKeyExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rateLimitResult := checkRateLimit(key.RateLimit, want.RateLimit)
		if rateLimitResult != nil {
			return rateLimitResult
		}

		return nil
	}
}

var testAccSentryKeyConfig = fmt.Sprintf(`
	resource "sentry_team" "test_team" {
		organization = "%s"
		name = "Test team"
	}

	resource "sentry_project" "test_project" {
		organization = "%s"
		team = "${sentry_team.test_team.id}"
		name = "Test project"
	}

	resource "sentry_key" "test_key" {
		organization = "%s"
		project = "${sentry_project.test_project.id}"
		name = "Test key"
	}
`, testOrganization, testOrganization, testOrganization)

var testAccSentryKeyUpdateConfig = fmt.Sprintf(`
	resource "sentry_team" "test_team" {
		organization = "%s"
		name = "Test team"
	}

	resource "sentry_project" "test_project" {
		organization = "%s"
		team = "${sentry_team.test_team.id}"
		name = "Test project"
	}

	resource "sentry_key" "test_key" {
		organization = "%s"
		project = "${sentry_project.test_project.id}"
		name = "Test key changed"
	}
`, testOrganization, testOrganization, testOrganization)

var testAccSentryRateLimitUpdateConfig = fmt.Sprintf(`
	resource "sentry_team" "test_team" {
		organization = "%s"
		name = "Test team"
	}

	resource "sentry_project" "test_project" {
		organization = "%s"
		team = "${sentry_team.test_team.id}"
		name = "Test project"
	}

	resource "sentry_key" "test_key" {
		organization = "%s"
		project = "${sentry_project.test_project.id}"
		name = "Test key changed"

		rate_limit_window = 300
		rate_limit_count = 2000
	}
`, testOrganization, testOrganization, testOrganization)
