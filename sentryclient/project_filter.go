package sentryclient

import (
	"encoding/json"
	"net/http"

	"github.com/dghubble/sling"
)

// ProjectFilter represents inbounding filters applied to a project
type ProjectFilter struct {
	Name string `json:"name,omitempty"`
}

// ProjectFilterService provides methods for accessing Sentry project filters
type ProjectFilterService struct {
	sling *sling.Sling
}

func newProjectFilterService(sling *sling.Sling) *ProjectFilterService {
	return &ProjectFilterService{
		sling: sling,
	}
}

// BrowserExtensionParams defines parameters for browser extension request
type BrowserExtensionParams struct {
	Active bool `json:"active"`
}

// UpdateBrowserExtensions updates configuration for browser extension filter
func (s *ProjectFilterService) UpdateBrowserExtensions(organizationSlug string, slug string, active bool) (*http.Response, error) {
	apiError := new(APIError)
	params := BrowserExtensionParams{Active: active}
	url := "projects/" + organizationSlug + "/" + slug + "/filters/browser-extensions/"
	resp, err := s.sling.New().Put(url).BodyJSON(params).Receive(nil, apiError)
	return resp, relevantError(err, *apiError)
}

// LegactBrowserParams defines parameters for legacy browser request
type LegactBrowserParams struct {
	Browsers []string `json:"subfilters"`
}

// UpdateLegacyBrowser updates configuration for legacy browser filters
func (s *ProjectFilterService) UpdateLegacyBrowser(organizationSlug string, slug string, browsers []string) (*http.Response, error) {
	apiError := new(APIError)
	params := LegactBrowserParams{Browsers: browsers}
	url := "projects/" + organizationSlug + "/" + slug + "/filters/legacy-browsers/"
	resp, err := s.sling.New().Put(url).BodyJSON(params).Receive(nil, apiError)
	return resp, relevantError(err, *apiError)
}

// FilterConfig represents configuration for project filter
type FilterConfig struct {
	BrowserExtension bool
	LegacyBrowsers   []string
}

// FilterConfigResponseItem represents an item in filter configuration response
type FilterConfigResponseItem struct {
	ID     string          `json:"id"`
	Active json.RawMessage `json:"active"`
}

// Get the filter configuration
func (s *ProjectFilterService) Get(organizationSlug string, slug string) (*FilterConfig, *http.Response, error) {
	apiError := new(APIError)
	filters := new([]FilterConfigResponseItem)
	url := "projects/" + organizationSlug + "/" + slug + "/filters/"
	resp, err := s.sling.New().Get(url).Receive(filters, apiError)
	filterConfig := &FilterConfig{
		BrowserExtension: false,
		LegacyBrowsers:   []string{},
	}

	for _, filter := range *filters {
		if filter.ID == "browser-extensions" && string(filter.Active) == "true" {
			filterConfig.BrowserExtension = true
		}
		if filter.ID == "legacy-browsers" && string(filter.Active) != "false" {
			var browsers []string
			jsonErr := json.Unmarshal(filter.Active, &browsers)
			if jsonErr != nil {
				panic(jsonErr)
			}
			filterConfig.LegacyBrowsers = browsers
		}
	}
	return filterConfig, resp, relevantError(err, *apiError)
}
