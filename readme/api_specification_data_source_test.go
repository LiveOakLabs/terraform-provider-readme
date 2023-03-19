package readme

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestAPISpecificationDataSource(t *testing.T) {
	defer gock.Off()

	// Mock the common API response.
	testResponse := `[{
		"id": "6398a4a594b26e00885e7ec0",
		"lastSynced": "2022-12-13T16:39:39.512Z",
		"category": {
			"id": "63f8dc63d70452003b73ff12",
			"title": "Test API Spec",
			"slug": "test-api-spec"
		},
		"source": "api",
		"title": "Test API Spec",
		"type": "oas",
		"version": "638cf4cfdea3ff0096d1a95a"
	}]`

	testCases := []struct {
		name        string
		config      string
		expectError string
		response    string
	}{
		// Unfiltered tests by ID or title.
		{
			name:   "it should return an API spec when a matching ID is found",
			config: `data "readme_api_specification" "test" { id = "6398a4a594b26e00885e7ec0" }`,
		},
		{
			name:   "it should return an API spec when a matching title is found",
			config: `data "readme_api_specification" "test" { title = "Test API Spec" }`,
		},
		{
			name: "it should return an API spec when both a matching ID and title are found",
			config: `data "readme_api_specification" "test" {
				id    = "6398a4a594b26e00885e7ec0"
				title = "Test API Spec"
			}`,
		},

		// Filter tests.
		{
			name: "it should return an API spec when a matching title and category ID is found",
			config: `data "readme_api_specification" "test" {
				title  = "Test API Spec"
				filter = { category_id = "63f8dc63d70452003b73ff12" }
			}`,
		},
		{
			name: "it should return an API spec when a matching title and category title is found",
			config: `data "readme_api_specification" "test" {
				title  = "Test API Spec"
				filter = { category_title = "Test API Spec" }
			}`,
		},
		{
			name: "it should return an API spec when a matching title and category slug is found",
			config: `data "readme_api_specification" "test" {
				title  = "Test API Spec"
				filter = { category_slug = "test-api-spec" }
			}`,
		},
		{
			name: "it should return an API spec when a matching title, category ID, and category title is found",
			config: `data "readme_api_specification" "test" {
				title  = "Test API Spec"
				filter = {
					category_id    = "63f8dc63d70452003b73ff12"
					category_title = "Test API Spec"
				}
			}`,
		},
		{
			name: "it should return an API spec when a matching title, category ID, and category slug is found",
			config: `data "readme_api_specification" "test" {
				title  = "Test API Spec"
				filter = {
					category_id  = "63f8dc63d70452003b73ff12"
					category_slug = "test-api-spec"
				}
			}`,
		},
		{
			name: "it should return an API spec when a matching id and category_id is found",
			config: `data "readme_api_specification" "test" {
				id     = "6398a4a594b26e00885e7ec0"
				filter = { category_id = "63f8dc63d70452003b73ff12" }
			}`,
		},
		{
			name: "it should return an API spec when a matching id and category_title is found",
			config: `data "readme_api_specification" "test" {
				id     = "6398a4a594b26e00885e7ec0"
				filter = { category_title = "Test API Spec" }
			}`,
		},
		{
			name: "it should return an API spec when a matching id and category_slug is found",
			config: `data "readme_api_specification" "test" {
				id     = "6398a4a594b26e00885e7ec0"
				filter = { category_slug = "test-api-spec" }
			}`,
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
			config: `data "readme_api_specification" "test" {
				id = "6398a4a594b26e00885e7ec0"
				filter = { category_slug = "does-not-exist" }
			}`,
			expectError: "Unable to find API specification with the specified criteria.",
		},
		{
			name: "it should return an error when no matching title and category slug is found",
			config: `data "readme_api_specification" "test" {
				title = "Test API Spec"
				filter = { category_slug = "does-not-exist" }
			}`,
			expectError: "Unable to find API specification with title: Test API Spec",
		},
		{
			name: "it should return an error when no matching title and category id is found",
			config: `data "readme_api_specification" "test" {
				title = "Test API Spec"
				filter = { category_id = "does-not-exist" }
			}`,
			expectError: "Unable to find API specification with title: Test API Spec",
		},
		{
			name: "it should return an error when no matching title and category title is found",
			config: `data "readme_api_specification" "test" {
				title = "Test API Spec"
				filter = { category_title = "does-not-exist" }
			}`,
			expectError: "Unable to find API specification with title: Test API Spec",
		},
		{
			name: "it should return an error when no matching title with a category is found",
			config: `data "readme_api_specification" "test" {
				title = "Test API Spec"
				filter = { has_category = true }
			}`,
			response: `[{
				"id": "6398a4a594b26e00885e7ec0",
				"lastSynced": "2022-12-13T16:39:39.512Z",
				"category": null,
				"source": "api",
				"title": "Test API Spec",
				"type": "oas",
				"version": "638cf4cfdea3ff0096d1a95a"
			}]`,
			expectError: "Unable to find API specification with title: Test API Spec",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := testResponse
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
							PreConfig: func() {
								gock.OffAll()
								gock.New(testURL).
									Get("/api-specification").
									MatchParam("perPage", "100").
									MatchParam("page", "1").
									Persist().
									Reply(200).
									SetHeaders(map[string]string{"link": `<>; rel="next", <>; rel="prev", <>; rel="last"`}).
									JSON(response)
							},
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
									"6398a4a594b26e00885e7ec0",
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"last_synced",
									"2022-12-13T16:39:39.512Z",
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"source",
									"api",
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"title",
									"Test API Spec",
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"type",
									"oas",
								),
								resource.TestCheckResourceAttr(
									"data.readme_api_specification.test",
									"version",
									"638cf4cfdea3ff0096d1a95a",
								),
							),
							PreConfig: func() {
								gock.OffAll()
								gock.New(testURL).
									Get("/api-specification").
									MatchParam("perPage", "100").
									MatchParam("page", "1").
									Persist().
									Reply(200).
									SetHeaders(map[string]string{"link": `<>; rel="next", <>; rel="prev", <>; rel="last"`}).
									JSON(response)
							},
						},
					},
				})
			}
		})
	}
}
