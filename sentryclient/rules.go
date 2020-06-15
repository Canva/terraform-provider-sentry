package sentryclient

import (
	"net/http"
	"time"
	"fmt"
	"encoding/json"
	"github.com/dghubble/sling"
	"log"
	// "github.com/mitchellh/mapstructure"

)

// Rule represents an alert rule configured for this project.
// https://github.com/getsentry/sentry/blob/9.0.0/src/sentry/api/serializers/models/rule.py
type Rule struct {
	ID          string          `json:"id"`
	ActionMatch string          `json:"actionMatch"`
	Environment string         `json:"environment"`
	Frequency   int             `json:"frequency"`
	Name        string          `json:"name"`
	Conditions  []RuleCondition `json:"conditions"`
	Actions     []RuleAction    `json:"actions"`
	Created     time.Time       `json:"dateCreated"`
}

// RuleCondition represents the conditions for each rule.
// https://github.com/getsentry/sentry/blob/9.0.0/src/sentry/api/serializers/models/rule.py
type RuleCondition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Attribute string `json:"attribute,omitempty"`
	Match     string `json:"match,omitempty"`
	Value     int `json:"value,omitempty"`
	Key       string `json:"key,omitempty"`
	Interval  string `json:"interval,omitempty"`
}

// RuleAction represents the actions will be taken for each rule based on its conditions.
// https://github.com/getsentry/sentry/blob/9.0.0/src/sentry/api/serializers/models/rule.py
type RuleAction struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Tags      string `json:"tags"`
	ChannelID string `json:"channel_id"`
	Channel   string `json:"channel"`
	Workspace string `json:"workspace"`
	Service   string `json:"service,omitempty"`
}

// ProjectKeyService provides methods for accessing Sentry project
// client key API endpoints.
// https://docs.sentry.io/api/projects/
type RuleService struct {
	sling *sling.Sling
}

func newRuleService(sling *sling.Sling) *RuleService {
	return &RuleService{
		sling: sling,
	}
}

// List alert rules configured for a project.
func (s *RuleService) List(organizationSlug string, projectSlug string) ([]Rule, *http.Response, error) {
	rules := new([]Rule)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("projects/"+organizationSlug+"/"+projectSlug+"/rules/").Receive(rules, apiError)
	// for i, rule := range (*rules) {
	// 	encoded_rule, _ := json.Marshal(rule)
	// 	log.Printf("rule[%d]: %s\n", i, string(encoded_rule))
	// }

	// for i := 0; i < len(*rules); i++ {
	// 	encoded_rule, _ := json.Marshal((*rules)[i])
	// 	log.Printf("rule[%i]: %s\n", i, string(encoded_rule))
	// }

	// log.Printf("LIST: %+v\n", rules)
	return *rules, resp, relevantError(err, *apiError)
}

func (s *RuleService) Read(organizationSlug string, projectSlug string, ruleId string) (Rule, *http.Response, error) {
	rule := new(Rule)
	apiError := new(APIError)
	resp, err := s.sling.New().Get("projects/"+organizationSlug+"/"+projectSlug+"/rules/"+ruleId+"/").Receive(rule, apiError)
	return *rule, resp, relevantError(err, *apiError)
}

// CreateRuleParams are the parameters for RuleService.Create.
type CreateRuleParams struct {
	ActionMatch string                       `json:"actionMatch"`
	Environment string                       `json:"environment,omitempty"`
	Frequency   int                          `json:"frequency"`
	Name        string                       `json:"name"`
	Conditions  []*CreateRuleConditionParams `json:"conditions"`
	Actions     []*CreateRuleActionParams    `json:"actions"`
}

// CreateRuleActionParams models the actions when creating the action for the rule.
type CreateRuleActionParams struct {
	ID        string `json:"id"`
	Tags      string `json:"tags"`
	Channel   string `json:"channel"`
	ChannelID string `json:"channel_id"`
	Workspace string `json:"workspace"`
	Action    string `json:"action,omitempty"`
	Service   string `json:"service,omitempty"`
}

// CreateRuleConditionParams models the conditions when creating the action for the rule.
type CreateRuleConditionParams struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Attribute string `json:"attribute,omitempty"`
	Match     string `json:"match,omitempty"`
	Value     int `json:"value,omitempty"`
	Key       string `json:"key,omitempty"`
	Interval  string `json:"interval,omitempty"`
}

// Create a new alert rule bound to a project.
func (s *RuleService) Create(organizationSlug string, projectSlug string, params *CreateRuleParams) (*Rule, *http.Response, error) {
	rule := new(Rule)
	apiError := new(APIError)
	resp, err := s.sling.New().Post("projects/"+organizationSlug+"/"+projectSlug+"/rules/").BodyJSON(params).Receive(rule, apiError)
	// log.Printf("EXPECTED: %+v\n", params)

	encoded_params, _ := json.Marshal(params)
	log.Printf("EXPECTED: %s\n", string(encoded_params))	

	// var decoded_params CreateRuleParams
	// mapstructure.WeakDecode(encoded_params, &decoded_params)
	// log.Printf("EXPECTED: %+v\n", decoded_params)


	fmt.Printf("HELLO WORLD")
	return rule, resp, relevantError(err, *apiError)
}

// Update a rule.
func (s *RuleService) Update(organizationSlug string, projectSlug string, ruleID string, params *Rule) (*Rule, *http.Response, error) {
	rule := new(Rule)
	apiError := new(APIError)
	resp, err := s.sling.New().Put("projects/"+organizationSlug+"/"+projectSlug+"/rules/"+ruleID+"/").BodyJSON(params).Receive(rule, apiError)
	return rule, resp, relevantError(err, *apiError)
}

// Delete a rule.
func (s *RuleService) Delete(organizationSlug string, projectSlug string, ruleID string) (*http.Response, error) {
	apiError := new(APIError)
	resp, err := s.sling.New().Delete("projects/"+organizationSlug+"/"+projectSlug+"/rules/"+ruleID+"/").Receive(nil, apiError)
	return resp, relevantError(err, *apiError)
}

