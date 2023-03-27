package readme

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

func TestAPISpecificationsDataSource(t *testing.T) {
	// Mock the default API response for all tests.
	testSpecs := []readme.APISpecification{
		{
			ID:         "6398a4a594b26e00885e7ec0",
			LastSynced: "2022-12-13T16:41:39.512Z",
			Category: readme.CategorySummary{
				ID:    "63f8dc63d70452003b73ff12",
				Title: "Test API Spec",
				Slug:  "test-api-spec",
			},
			Source:  "api",
			Title:   "Test API Spec",
			Type:    "oas",
			Version: "638cf4cfdea3ff0096d1a95a",
		},
		{
			ID:         "6398a4a594b26e00885e7ec1",
			LastSynced: "2022-12-13T16:40:39.512Z",
			Category: readme.CategorySummary{
				ID:    "63f8dc63d70452003b73ff13",
				Title: "Another Test API Spec",
				Slug:  "another-test-api-spec",
			},
			Source:  "api",
			Title:   "Another Test API Spec",
			Type:    "oas",
			Version: "638cf4cfdea3ff0096d1a95b",
		},
		{
			ID:         "6398a4a594b26e00885e7ec2",
			LastSynced: "2022-12-13T16:39:39.512Z",
			Source:     "api",
			Title:      "Test API Spec Without Category",
			Type:       "oas",
			Version:    "638cf4cfdea3ff0096d1a95c",
		},
	}

	testCases := []struct {
		name         string                    // The name of the test.
		config       string                    // The Terraform config to use for the test.
		expectError  string                    // If set, the test will expect an error matching this regex to be returned.
		response     []readme.APISpecification // List of specs that should be returned.
		responseCode int                       // Expected HTTP response code. Defaults to 200.
		excluded     []readme.APISpecification // List of specs that should explicitly *not* be returned.
	}{
		{
			name:     "it should return a list of all API specs when no filters are provided",
			config:   `data "readme_api_specifications" "test" {}`,
			response: testSpecs,
		},

		// === Sorting.
		{
			name:     "it should return a list of unsorted API specs when sort_by is not provided",
			config:   `data "readme_api_specifications" "test" {}`,
			response: []readme.APISpecification{testSpecs[0], testSpecs[1], testSpecs[2]},
		},
		{
			name: "it should return a list of API specs sorted by title in ascending order when sort_by=title is provided",
			config: `data "readme_api_specifications" "test" {
				sort_by = "title"
			}`,
			response: []readme.APISpecification{testSpecs[1], testSpecs[0], testSpecs[2]},
		},
		{
			name: "it should return a list of API specs sorted by last_synced in ascending order when sort_by=last_synced is provided",
			config: `data "readme_api_specifications" "test" {
				sort_by = "last_synced"
				}`,
			response: []readme.APISpecification{testSpecs[2], testSpecs[1], testSpecs[0]},
		},

		// ==== Filter by spec title.
		{
			name: "it should return API specs that match the provided title filter",
			config: `data "readme_api_specifications" "test" {
				filter = { title = ["Test API Spec", "Another Test API Spec"] }
			}`,
			response: []readme.APISpecification{testSpecs[0], testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[2]},
		},

		// ==== Filter by spec version ID.
		{
			name: "it should return API specs that match the provided version filter",
			config: `data "readme_api_specifications" "test" {
				filter = { version = ["638cf4cfdea3ff0096d1a95a", "638cf4cfdea3ff0096d1a95b"] }
			}`,
			response: []readme.APISpecification{testSpecs[0], testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[2]},
		},

		// ==== Filter by has_category.
		{
			name: "it should return API specs that match the provided has_category filter",
			config: `data "readme_api_specifications" "test" {
				filter = { has_category = true }
			}`,
			response: []readme.APISpecification{testSpecs[0], testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[2]},
		},
		{
			name: "it should return only API specs that have a category when has_category is true and used with another filter",
			config: `data "readme_api_specifications" "test" {
				filter = {
					has_category = true
					category_slug = [
						"test-api-spec",
						"another-test-api-spec",
						"test-api-spec-without-category"
					]
				}
			}`,
			response: []readme.APISpecification{testSpecs[0], testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[2]},
		},
		{
			name: "it should return an empty list when has_category is true and used with another filter that doesn't match any API specs",
			config: `data "readme_api_specifications" "test" {
				filter = {
					has_category = true
					category_slug = ["test-api-spec-without-category"]
				}
			}`,
			response: []readme.APISpecification{},
		},

		// ==== Filter by category title.
		{
			name: "it should return API specs that match the provided category_title filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_title = ["Test API Spec"] }
			}`,
			response: []readme.APISpecification{testSpecs[0]},
			excluded: []readme.APISpecification{testSpecs[1], testSpecs[2]},
		},

		// ==== Filter by category ID.
		{
			name: "it should return API specs that match the provided category_id filter when has_category is not provided",
			config: `data "readme_api_specifications" "test" {
				filter = { category_id = ["63f8dc63d70452003b73ff13"] }
			}`,
			response: []readme.APISpecification{testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[0], testSpecs[2]},
		},
		{
			name: "it should return API specs that match the provided category_id filter when has_category is true",
			config: `data "readme_api_specifications" "test" {
				filter = {
					has_category = true
					category_id  = [
						"63f8dc63d70452003b73ff12",
						"63f8dc63d70452003b73ff13"
					]
				}
			}`,
			response: []readme.APISpecification{testSpecs[0]},
			excluded: []readme.APISpecification{testSpecs[1], testSpecs[2]},
		},

		// ==== Filter by category slug.
		{
			name: "it should return API specs that match the provided category_slug filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_slug = ["another-test-api-spec"] }
			}`,
			response: []readme.APISpecification{testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[0], testSpecs[2]},
		},

		// ==== Filter combinations.
		{
			name: "it should return API specs that match the provided category_id and category_title filters",
			config: `data "readme_api_specifications" "test" {
				filter = {
					category_id    = ["6398a4a594b26e00885e7ec2"]
					category_title = ["Another Test API Spec"]
					has_category   = true
				}
			}`,
			response: []readme.APISpecification{testSpecs[1]},
			excluded: []readme.APISpecification{testSpecs[0], testSpecs[2]},
		},
		{
			name: "it does not return specs that don't match any filters provided",
			config: `data "readme_api_specifications" "test" {
				filter = {
					has_category   = true
					category_id    = ["000000011111122222223333"]
					category_slug  = ["this-doesnt-exist"]
					category_title = ["This doesn't exist"]
					title          = ["This doesn't exist either"]
				}
			}`,
			response: []readme.APISpecification{},
		},
		{
			name: "it does not return API specs that don't match the category_slug filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_slug = ["test-api-spec"] }
			}`,
			response: []readme.APISpecification{},
		},
		{
			name: "it does not return API specs that don't match the category_title filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_title = ["This does not exist"] }
			}`,
			response: []readme.APISpecification{},
		},

		// ==== Error handling.
		{
			name: "it should return an error when the API response is invalid",
			config: `data "readme_api_specifications" "test" {
				filter = { category_title = ["Test API Spec"] }
			}`,
			response:     []readme.APISpecification{},
			responseCode: 400,
			expectError:  "Unable to retrieve API specifications",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := testSpecs
			if tc.response != nil {
				response = tc.response
			}

			responseCode := 200
			if tc.responseCode != 0 {
				responseCode = tc.responseCode
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
									Reply(responseCode).
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
								apiSpecificationsChecks(t, response, tc.excluded)...,
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

// apiSpecificationsChecks returns a list of TestCheckFuncs to test the data source.
// It ensures the number of specs returned is correct, the excluded specs are not returned,
// and the specs returned are correct.
func apiSpecificationsChecks(t *testing.T, specs, excluded []readme.APISpecification) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	// Ensure the number of specs returned is correct.
	checks = append(checks, resource.TestCheckResourceAttr(
		"data.readme_api_specifications.test",
		"specs.#",
		strconv.Itoa(len(specs)),
	))

	// Ensure the excluded specs are not returned.
	for _, exclude := range excluded {
		for _, spec := range specs {
			if spec.ID == exclude.ID {
				t.Fatalf("Expected to exclude API specification with ID: %s", spec.ID)
			}
		}
	}

	// Ensure the specs returned are correct.
	for i, spec := range specs {
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.id", i),
			spec.ID,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.last_synced", i),
			spec.LastSynced,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.source", i),
			spec.Source,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.title", i),
			spec.Title,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.type", i),
			spec.Type,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.version", i),
			spec.Version,
		))
	}

	return checks
}
