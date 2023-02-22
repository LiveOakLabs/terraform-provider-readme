package readme

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// NewTest sets up the provider for testing.
func NewTest(version string) provider.Provider {
	return &readmeProvider{Version: "test"}
}

const (
	// testURL is a dummy URL the provider is configured with and the mock HTTP
	// service responds to.
	testURL = "http://testing/api/v1"
	// testToken is a dummy token the provider is configured with and used
	// throughout tests.
	testToken = "hunter2"
	// providerConfig is a shared configuration that sets a mock url and token.
	// The URL points to our gock mock server.
	providerConfig = (`
		provider "readme" {
			api_token = "` + testToken + `"
			api_url   = "` + testURL + `"
		}
	`)
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// testing. The factory function will be invoked for every Terraform CLI command
// executed to create a provider server to which the CLI can reattach.
var testProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"readme": providerserver.NewProtocol6WithError(NewTest("test")),
}

// mockAPIError represents a generic ReadMe API error response for use throughout tests.
var mockAPIError = readme.APIErrorResponse{
	Error:      "TEST_ERROR",
	Message:    "This is a test error message response",
	Suggestion: "",
	Docs:       "",
	Help:       "",
	Poem:       []string{"one"},
}

// removeIndents is a helper for removing tabs from indented heredocs in multi-line strings.
func removeIndents(str string) string {
	return strings.ReplaceAll(str, "\t", "")
}

// replaceNewlines is a helper for replacing newlines with a `\n` literal to create a single
// line string from a multi-line string.
func replaceNewlines(str string) string {
	return strings.ReplaceAll(str, "\n", `\n`)
}

func TestProvider_MissingAPIToken(t *testing.T) {
	expectError, _ := regexp.Compile(`Missing ReadMe API Token`)

	invalidProviderConfig := `
	provider "readme" {
		api_token = ""
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidProviderConfig + `data "readme_project" "test" {}`,
				ExpectError: expectError,
			},
		},
	})
}

func TestProvider_EmptyAPIURL(t *testing.T) {
	expectError, _ := regexp.Compile(`Missing ReadMe API URL`)

	invalidProviderConfig := `
	provider "readme" {
		api_token = "hunter2"
		api_url   = ""
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidProviderConfig + `data "readme_project" "test" {}`,
				ExpectError: expectError,
			},
		},
	})
}
