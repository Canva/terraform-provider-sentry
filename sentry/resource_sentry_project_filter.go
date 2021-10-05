package sentry

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/canva/terraform-provider-sentry/sentryclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSentryFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceSentryFilterUpdate,
		Read:   resourceSentryFilterRead,
		Update: resourceSentryFilterUpdate,
		Delete: resourceSentryFilterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSentryFilterImporter,
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
			"browser_extension": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether to filter out events from browser extension",
			},
			"legacy_browsers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Events from these legacy browsers will be ignored",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceSentryFilterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)
	org := d.Get("organization").(string)
	project := d.Get("project").(string)
	filterConfig, _, err := client.ProjectFilter.Get(org, project)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s-%s_filter", org, project))
	d.Set("organization", org)
	d.Set("project", project)
	d.Set("browser_extension", filterConfig.BrowserExtension)
	return nil
}

func resourceSentryFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)
	browserExtension := d.Get("browser_extension").(bool)
	inputLegacyBrowsers := d.Get("legacy_browsers").([]interface{})
	legacyBrowsers := make([]string, len(inputLegacyBrowsers))
	for idx, browser := range inputLegacyBrowsers {
		legacyBrowsers[idx] = browser.(string)
	}
	org := d.Get("organization").(string)
	project := d.Get("project").(string)
	_, err := client.ProjectFilter.UpdateBrowserExtensions(org, project, browserExtension)
	if err != nil {
		return err
	}
	_, err = client.ProjectFilter.UpdateLegacyBrowser(org, project, legacyBrowsers)
	if err != nil {
		return err
	}
	return resourceSentryFilterRead(d, meta)
}

func resourceSentryFilterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*sentryclient.Client)
	org := d.Get("organization").(string)
	project := d.Get("project").(string)
	_, err := client.ProjectFilter.UpdateBrowserExtensions(org, project, false)
	if err != nil {
		return err
	}
	_, err = client.ProjectFilter.UpdateLegacyBrowser(org, project, []string{})
	if err != nil {
		return err
	}
	return resourceSentryFilterRead(d, meta)
}

func resourceSentryFilterImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	addrID := d.Id()

	log.Printf("[DEBUG] Importing filter using ADDR ID %s", addrID)

	parts := strings.Split(addrID, "/")

	if len(parts) != 3 {
		return nil, errors.New("Project import requires an ADDR ID of the following schema org-slug/project-slug/rule-id")
	}

	d.Set("organization", parts[0])
	d.Set("project", parts[1])
	d.SetId(parts[2])

	return []*schema.ResourceData{d}, nil
}
