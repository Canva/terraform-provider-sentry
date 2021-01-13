package sentryclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readRequestBody(r *http.Request) string {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}
	str := string(b)
	str = strings.TrimSuffix(str, "\n")
	return str
}

func TestBrowserExtensionFilter(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/api/0/projects/test_org/test_project/filters/browser-extensions/", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "PUT", r)
		body := readRequestBody(r)
		assert.Equal(t, body, `{"active":true}`)
		w.Header().Set("Content-Type", "application/json")
	})
	client := NewClient(httpClient, nil, "")
	_, err := client.ProjectFilter.UpdateBrowserExtensions("test_org", "test_project", true)
	assert.NoError(t, err)
}

func TestLegacyBrowserFilter(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/api/0/projects/test_org/test_project/filters/legacy-browsers/", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "PUT", r)
		body := readRequestBody(r)
		assert.Equal(t, body, `{"subfilters":["ie_pre_9","ie10"]}`)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "")
	})
	client := NewClient(httpClient, nil, "")
	browsers := []string{"ie_pre_9", "ie10"}
	_, err := client.ProjectFilter.UpdateLegacyBrowser("test_org", "test_project", browsers)
	assert.NoError(t, err)
}

func TestGetWithLegacyExtension(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/api/0/projects/test_org/test_project/filters/", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "GET", r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":"browser-extensions","active":false,"description":"Certain browser extensions will inject inline scripts and are known to cause errors.","name":"Filter out errors known to be caused by browser extensions","hello":"browser-extensions - Filter out errors known to be caused by browser extensions"},{"id":"localhost","active":false,"description":"This applies to both IPv4 addresses.","name":"Filter out events coming from localhost","hello":"localhost - Filter out events coming from localhost"},{"id":"legacy-browsers","active":["ie_pre_9"],"description":"Older browsers often give less accurate information, and while they may report valid issues, the context to understand them is incorrect or missing.","name":"Filter out known errors from legacy browsers","hello":"legacy-browsers - Filter out known errors from legacy browsers"},{"id":"web-crawlers","active":true,"description":"Some crawlers may execute pages in incompatible ways which then cause errors that are unlikely to be seen by a normal user.","name":"Filter out known web crawlers","hello":"web-crawlers - Filter out known web crawlers"}]`)
	})
	client := NewClient(httpClient, nil, "")
	filterConfig, _, err := client.ProjectFilter.Get("test_org", "test_project")
	assert.NoError(t, err)
	expectedConfig := FilterConfig{
		LegacyBrowsers:   []string{"ie_pre_9"},
		BrowserExtension: false,
	}
	assert.Equal(t, *filterConfig, expectedConfig)
}

func TestGetWithoutLegacyExtension(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/api/0/projects/test_org/test_project/filters/", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "GET", r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":"browser-extensions","active":true,"description":"Certain browser extensions will inject inline scripts and are known to cause errors.","name":"Filter out errors known to be caused by browser extensions","hello":"browser-extensions - Filter out errors known to be caused by browser extensions"},{"id":"localhost","active":false,"description":"This applies to both IPv4  addresses.","name":"Filter out events coming from localhost","hello":"localhost - Filter out events coming from localhost"},{"id":"legacy-browsers","active":false,"description":"Older browsers often give less accurate information, and while they may report valid issues, the context to understand them is incorrect or missing.","name":"Filter out known errors from legacy browsers","hello":"legacy-browsers - Filter out known errors from legacy browsers"},{"id":"web-crawlers","active":true,"description":"Some crawlers may execute pages in incompatible ways which then cause errors that are unlikely to be seen by a normal user.","name":"Filter out known web crawlers","hello":"web-crawlers - Filter out known web crawlers"}]`)
	})
	client := NewClient(httpClient, nil, "")
	filterConfig, _, err := client.ProjectFilter.Get("test_org", "test_project")
	assert.NoError(t, err)
	expectedConfig := FilterConfig{
		LegacyBrowsers:   []string{},
		BrowserExtension: true,
	}
	assert.Equal(t, *filterConfig, expectedConfig)
}
