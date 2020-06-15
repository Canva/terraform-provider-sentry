package sentry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/canva/terraform-provider-sentry/sentryclient"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/google/go-cmp/cmp"
)

const projectSlug = "test-project"

func TestAccSentryRule_basic(t *testing.T) {
	var rule sentryclient.Rule


	testAccSentryRuleUpdateConfig := fmt.Sprintf(`
	resource "sentry_rule" "test_rule" {
		name = "Important Issue"
		organization = "%s"
		project = "%s"
		action_match = "all"
		frequency    = 1300
		environment  = "prod"
		actions = [
			{
			id = "sentry.rules.actions.notify_event.NotifyEventAction"
			}
		]
		conditions = [
			{
				id = "sentry.rules.conditions.event_frequency.EventFrequencyCondition"
				value = 101
				name = "An issue is seen more than 100 times in 1 hour"
				interval = "1h"
		 	},
		 	{
				id = "sentry.rules.conditions.event_frequency.EventUniqueUserFrequencyCondition"
				interval = "1m"
				name = "An issue is seen by more than 25 users in 1 minute"
				value = 30
		 	}
		]
	}
	`, testOrganization, projectSlug)

	fmt.Printf("%s", testAccSentryRuleUpdateConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSentryRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSentryRuleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSentryRuleExists("sentry_rule.test_rule", &rule, testOrganization, projectSlug),
					testAccCheckSentryRuleAttributes(&rule, &testAccSentryRuleExpectedAttributes{
						Name: "Important Issue",
						// Organization: testOrganization,
						// Project: projectSlug,
						ActionMatch: "all",
						Frequency: 1440,
						Environment: "prod",
						Actions: []sentryclient.RuleAction{
							{
								ID: "sentry.rules.actions.notify_event.NotifyEventAction",
								Name: "Send a notification (for all legacy integrations)",  // Default name added by Sentry
							},
						},
						Conditions: []sentryclient.RuleCondition{
							{
								ID: "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
								Value: 100,
								Name: "An issue is seen more than 100 times in 1m",
								Interval: "1m",
							},
							{
								ID: "sentry.rules.conditions.event_frequency.EventUniqueUserFrequencyCondition",
								Interval: "1m",
								Name: "An issue is seen by more than 25 users in 1m",
								Value: 25,
							},
						},
					}),
				),
			},
			{
				Config: testAccSentryRuleUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSentryRuleExists("sentry_rule.test_rule", &rule, testOrganization, projectSlug),
					testAccCheckSentryRuleAttributes(&rule, &testAccSentryRuleExpectedAttributes{
						Name: "Important Issue",
						// Organization: testOrganization,
						// Project: projectSlug,
						ActionMatch: "all",
						Frequency: 1300,
						Environment: "prod",
						Actions: []sentryclient.RuleAction{
							{
								ID: "sentry.rules.actions.notify_event.NotifyEventAction",
								Name: "Send a notification (for all legacy integrations)",  // Default name added by Sentry
							},
						},
						Conditions: []sentryclient.RuleCondition{
							{
								ID: "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
								Value: 101,
								Name: "An issue is seen more than 100 times in 1m",
								Interval: "1h",
							},
							{
								ID: "sentry.rules.conditions.event_frequency.EventUniqueUserFrequencyCondition",
								Interval: "1m",
								Name: "An issue is seen by more than 25 users in 1m",
								Value: 30,
							},
						},
					}),
				),
			},
		},
	})
}

func testAccCheckSentryRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*sentryclient.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sentry_team" {
			continue
		}

		team, resp, err := client.Teams.Get(
			rs.Primary.Attributes["organization"],
			rs.Primary.ID,
		)
		if err == nil {
			if team != nil {
				return errors.New("Team still exists")
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckSentryRuleExists(n string, rule *sentryclient.Rule, org string, proj string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Rule ID set")
		}

		client := testAccProvider.Meta().(*sentryclient.Client)
		SentryRules, _, err := client.Rules.List(org, proj)

		if err != nil {
			return err
		}
		fmt.Printf("EXPECTED: %+v\n", rs.Primary)

		var SentryRule *sentryclient.Rule

		for _, r := range SentryRules {
			fmt.Printf("RULE: %+v\n", r)
			if r.ID == rs.Primary.ID {
				SentryRule = &r
				break
			}
		}
		if SentryRule == nil {
			return errors.New("Could not find Rule.")
		}

		*rule = *SentryRule
		return nil
	}
}

type testAccSentryRuleExpectedAttributes struct {
	Name string
	// Organization string
	// Project string
	ActionMatch string
	Frequency int
	Environment string
	Actions []sentryclient.RuleAction
	Conditions []sentryclient.RuleCondition
}



func testAccCheckSentryRuleAttributes(rule *sentryclient.Rule, want *testAccSentryRuleExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if rule.Name != want.Name {
			return fmt.Errorf("got rule name %q; want %q", rule.Name, want.Name)
		}
		// if rule.Organization != want.Organization {
		// 	return fmt.Errorf("got organization %q; want %q", rule.Organization, want.Organization)
		// }
		// if rule.Project != want.Project {
		// 	return fmt.Errorf("got project %q; want %q", rule.Project, want.Project)
		// }

		if rule.ActionMatch != want.ActionMatch {
			return fmt.Errorf("got action_match %s; want %s", rule.ActionMatch, want.ActionMatch)
		}

		if rule.Frequency != want.Frequency {
			return fmt.Errorf("got frequency %d; want %d", rule.Frequency, want.Frequency)
		}

		if rule.Environment != want.Environment {
			return fmt.Errorf("got environment %s; want %s", rule.Environment, want.Environment)
		}

		if !cmp.Equal(rule.Actions, want.Actions){
			return fmt.Errorf("got actions: %+v\n; want %+v\n", rule.Actions, want.Actions)	
		}

		if !cmp.Equal(rule.Conditions, want.Conditions){
			return fmt.Errorf("got conditions: %+v\n; want %+v\n", rule.Conditions, want.Conditions)	
		}
		// 	return errors.New("got empty slug; want non-empty slug")
		// }

		// if want.Slug != "" && team.Slug != want.Slug {
		// 	return fmt.Errorf("got slug %q; want %q", team.Slug, want.Slug)
		// }

		return nil
	}
}


var testAccSentryRuleConfig = fmt.Sprintf(`
resource "sentry_rule" "test_rule" {
	name = "Important Issue"
	organization = "%s"
	project = "%s"
	action_match = "all"
	frequency    = 1440
	environment  = "prod"
	actions = [
		{
			id = "sentry.rules.actions.notify_event.NotifyEventAction"
		}
	]
	conditions = [
		{
			id = "sentry.rules.conditions.event_frequency.EventFrequencyCondition"
			value = 100
			name = "An issue is seen more than 100 times in 1m"
			interval = "1m"
	 	},
	 	{
			id = "sentry.rules.conditions.event_frequency.EventUniqueUserFrequencyCondition"
			interval = "1m"
			name = "An issue is seen by more than 25 users in 1m"
			value = 25
	 	}
	]
}
`, testOrganization, projectSlug)
