package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"github.com/liveoaklabs/terraform-provider-readme/internal/testdata"
	"gopkg.in/h2non/gock.v1"
)

func TestAPISpecificationDataSource(t *testing.T) {
	defer gock.Off()
	testResponse := testdata.APISpecifications[:1]

	testCases := []struct {
		name        string
		config      string
		expectError string
		response    string
	}{
		// Unfiltered tests by ID or title.
		{
			name: "it should return an API spec when a matching ID is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					id = "%s"
				}`,
				testResponse[0].ID,
			),
		},
		{
			name: "it should return an API spec when a matching title is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title = "%s"
				}`,
				testResponse[0].Title,
			),
		},
		{
			name: "it should return an API spec when both a matching ID and title are found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					id    = "%s"
					title = "%s"
				}`,
				testResponse[0].ID,
				testResponse[0].Title,
			),
		},

		// Filter tests.
		{
			name: "it should return an API spec when a matching title and category ID is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { category_id = "%s" }
				}`,
				testResponse[0].Title,
				testResponse[0].Category.ID,
			),
		},
		{
			name: "it should return an API spec when a matching title and category title is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { category_title = "%s" }
				}`,
				testResponse[0].Title,
				testResponse[0].Category.Title,
			),
		},
		{
			name: "it should return an API spec when a matching title and category slug is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { category_slug = "%s" }
				}`,
				testResponse[0].Title,
				testResponse[0].Category.Slug,
			),
		},
		{
			name: "it should return an API spec when a matching title, category ID, and category title is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = {
						category_id    = "%s"
						category_title = "%s"
					}
				}`,
				testResponse[0].Title,
				testResponse[0].Category.ID,
				testResponse[0].Category.Title,
			),
		},
		{
			name: "it should return an API spec when a matching title, category ID, and category slug is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = {
						category_id   = "%s"
						category_slug = "%s"
					}
				}`,
				testResponse[0].Title,
				testResponse[0].Category.ID,
				testResponse[0].Category.Slug,
			),
		},
		{
			name: "it should return an API spec when a matching id and category_id is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					id     = "%s"
					filter = { category_id = "%s" }
				}`,
				testResponse[0].ID,
				testResponse[0].Category.ID,
			),
		},
		{
			name: "it should return an API spec when a matching id and category_title is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					id     = "%s"
					filter = { category_title = "%s" }
				}`,
				testResponse[0].ID,
				testResponse[0].Category.Title,
			),
		},
		{
			name: "it should return an API spec when a matching id and category_slug is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					id     = "%s"
					filter = { category_slug = "%s" }
				}`,
				testResponse[0].ID,
				testResponse[0].Category.Slug,
			),
		},

		// Negative tests
		{
			name:        "it should return an error when no ID or title is provided",
			config:      `data "readme_api_specification" "test" { }`,
			expectError: "An ID or title must be specified to retrieve an API specification.",
		},
		{
			name:        "it should return an error when no API specs exist in a project",
			config:      `data "readme_api_specification" "test" { title = "Test API Spec" }`,
			expectError: "No API specifications were found in the ReadMe project",
			response:    `[]`,
		},
		{
			name: "it should return an error when no matching ID is found",
			config: `data "readme_api_specification" "test" {
				id = "does-not-exist"
			}`,
			expectError: "API specification not found",
		},
		{
			name: "it should return an error when no matching title is found",
			config: `data "readme_api_specification" "test" {
				title = "does-not-exist"
			}`,
			expectError: "Unable to find API specification with title: does-not-exist",
		},
		{
			name: "it should return an error when no matching ID and category slug is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					id     = "%s"
					filter = { category_slug = "does-not-exist" }
				}`,
				testResponse[0].ID,
			),
			expectError: "Unable to find API specification with the specified criteria.",
		},
		{
			name: "it should return an error when no matching title and category slug is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { category_slug = "does-not-exist" }
				}`,
				testResponse[0].Title,
			),
			expectError: "Unable to find API specification with title: Test API Spec",
		},
		{
			name: "it should return an error when no matching title and category id is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { category_id = "does-not-exist" }
				}`,
				testResponse[0].Title,
			),
			expectError: "Unable to find API specification with title: Test API Spec",
		},
		{
			name: "it should return an error when no matching title and category title is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { category_title = "does-not-exist" }
				}`,
				testResponse[0].Title,
			),
			expectError: "Unable to find API specification with title: Test API Spec",
		},
		{
			name: "it should return an error when no matching title with a category is found",
			config: fmt.Sprintf(
				`data "readme_api_specification" "test" {
					title  = "%s"
					filter = { has_category = false }
				}`,
				testResponse[0].Title,
			),
			response: testdata.ToJSON(
				[]readme.APISpecification{testdata.APISpecificationsNoCategory},
			),
			expectError: "Unable to find API specification with title: Test API Spec",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := testdata.ToJSON(testResponse)
			if tc.response != "" {
				response = tc.response
			}

			if tc.expectError != "" {
				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV6ProviderFactories: testProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config:      providerConfig + tc.config,
							ExpectError: regexp.MustCompile(tc.expectError),
							PreConfig:   testdata.APISpecificationRespond(response, 200),
						},
					},
				})
			} else {
				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV6ProviderFactories: testProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: providerConfig + tc.config,
							Check: resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"id",
									testdata.APISpecifications[0].ID,
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"last_synced",
									testdata.APISpecifications[0].LastSynced,
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"source",
									testdata.APISpecifications[0].Source,
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"title",
									testdata.APISpecifications[0].Title,
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"type",
									testdata.APISpecifications[0].Type,
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"version",
									testdata.APISpecifications[0].Version,
								),
							),
							PreConfig: testdata.APISpecificationRespond(response, 200),
						},
					},
				})
			}
		})
	}
}
