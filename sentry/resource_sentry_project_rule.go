package sentry

import (
	"errors"
	"encoding/json"
	"log"

	"github.com/canva/terraform-provider-sentry/sentryclient"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mitchellh/mapstructure"
)

const (
	defaultActionMatch = "any"
	defaultFrequency   = 30
)

func resourceSentryRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceSentryRuleCreate,
		Read:   resourceSentryRuleRead,
		Update: resourceSentryRuleUpdate,
		Delete: resourceSentryRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSentryRuleImporter,
		},

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the organization the project belongs to",
			},
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the project to create the plugin for",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The rule name",
			},
			"action_match": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"actions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"conditions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"frequency": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Perform actions at most once every X minutes",
			},
			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Perform rule in a specific environment",
			},
		},
	}
}

func resourceSentryRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)

	name := d.Get("name").(string)
	org := d.Get("organization").(string)
	project := d.Get("project").(string)
	environment := d.Get("environment").(string)
	actionMatch := d.Get("action_match").(string)
	inputConditions := d.Get("conditions").([]interface{})
	inputActions := d.Get("actions").([]interface{})
	frequency := d.Get("frequency").(int)

	if actionMatch == "" {
		actionMatch = defaultActionMatch
	}
	if frequency == 0 {
		frequency = defaultFrequency
	}

	conditions := make([]*sentryclient.CreateRuleConditionParams, len(inputConditions))
	for i, ic := range inputConditions {
		var condition sentryclient.CreateRuleConditionParams
		mapstructure.Decode(ic, &condition)
		conditions[i] = &condition
	}
	actions := make([]*sentryclient.CreateRuleActionParams, len(inputActions))
	for i, ia := range inputActions {
		var action sentryclient.CreateRuleActionParams
		mapstructure.Decode(ia, &action)
		actions[i] = &action
	}

	params := &sentryclient.CreateRuleParams{
		ActionMatch: actionMatch,
		Environment: environment,
		Frequency:   frequency,
		Name:        name,
		Conditions:  conditions,
		Actions:     actions,
	}

	if environment != "" {
		params.Environment = environment
	}

	rule, _, err := client.Rules.Create(org, project, params)
	if err != nil {
		return err
	}

	encoded_params, _ := json.Marshal(params)
	log.Printf("create rule params: %s\n", string(encoded_params))

	encoded_rule, _ := json.Marshal(rule)
	log.Printf("rule create rule: %s\n", string(encoded_rule))


	d.SetId(rule.ID)

	return resourceSentryRuleRead(d, meta)
}

func resourceSentryRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)
	org := d.Get("organization").(string)
	project := d.Get("project").(string)
	id := d.Id()

	// rules, resp, err := client.Rules.List(org, project)
	// if found, err := checkClientGet(resp, err, d); !found {
	// 	return err
	// }

	rules, _, err := client.Rules.List(org, project)
	if err != nil {
		d.SetId("")
		return nil
	}

	var rule *sentryclient.Rule
	for _, r := range rules {
		if r.ID == id {
			rule = &r
			break
		}
	}

	if rule == nil {
		return errors.New("Could not find rule with ID " + id)
	}

	encoded_rule, _ := json.Marshal(rule)
	log.Printf("rule read: %s\n", string(encoded_rule))

	d.SetId(rule.ID)
	d.Set("name", rule.Name)
	d.Set("actions", rule.Actions)
	d.Set("conditions", rule.Conditions)
	d.Set("frequency", rule.Frequency)
	d.Set("environment", rule.Environment)

	d.Set("action_match", rule.ActionMatch)
	d.Set("organization", org)
	d.Set("project", project)

	return nil
}

func resourceSentryRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)

	id := d.Id()
	name := d.Get("name").(string)
	org := d.Get("organization").(string)
	project := d.Get("project").(string)
	environment := d.Get("environment").(string)
	actionMatch := d.Get("action_match").(string)
	inputConditions := d.Get("conditions").([]interface{})
	inputActions := d.Get("actions").([]interface{})
	frequency := d.Get("frequency").(int)

	if actionMatch == "" {
		actionMatch = defaultActionMatch
	}
	if frequency == 0 {
		frequency = defaultFrequency
	}

	conditions := make([]sentryclient.RuleCondition, len(inputConditions))
	for i, ic := range inputConditions {
		var condition sentryclient.RuleCondition
		mapstructure.Decode(ic, &condition)
		conditions[i] = condition
	}
	actions := make([]sentryclient.RuleAction, len(inputActions))
	for i, ia := range inputActions {
		var action sentryclient.RuleAction
		mapstructure.Decode(ia, &action)
		actions[i] = action
	}

	params := &sentryclient.Rule{
		ID:          id,
		ActionMatch: actionMatch,
		Environment: environment,
		Frequency:   frequency,
		Name:        name,
		Conditions:  conditions,
		Actions:     actions,
	}

	if environment != "" {
		params.Environment = environment
	}

	_, _, err := client.Rules.Update(org, project, id, params)
	if err != nil {
		return err
	}

	return resourceSentryRuleRead(d, meta)
}

func resourceSentryRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)

	id := d.Id()
	org := d.Get("organization").(string)
	project := d.Get("project").(string)

	_, err := client.Rules.Delete(org, project, id)
	return err
}
