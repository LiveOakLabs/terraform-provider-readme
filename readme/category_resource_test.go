package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

// mockCategoryCreate is a common response to a category creation. Creating a
// category returns a unique response - the 'version' key is a full dictionary
// of version info. All other category responses, including updating, returns
// the same struct with the 'version' key a string of the version ID.
//
// Refer to the 'category_data_source' file for the common mockCategory shared
// throughout category tests.
var mockCategoryCreate = readme.CategorySaved{
	CreatedAt: "2023-01-06T21:25:39.543Z",
	ID:        "63b891d3ee384600680cea03",
	Order:     9999,
	Project:   "63b891d3ee384600680ce9eb",
	Reference: false,
	Slug:      "documentation",
	Title:     "Documentation",
	Type:      "guide",
	Version:   mockVersion,
}

// createCategoryTestStep is a helper function that returns a resource test step
// that creates a category successfully.
// A 'check' parameter accepts a `resource.TestCheckFunc`, which may also be nil
// for no checks.
func createCategoryTestStep(check resource.TestCheckFunc) resource.TestStep {
	return resource.TestStep{
		Config: providerConfig + `resource "readme_category" "test" {
			title = "` + mockCategoryCreate.Title + `"
			type  = "` + mockCategoryCreate.Type + `"
		}`,
		PreConfig: func() {
			// Mock the request to create the resource.
			gock.New(testURL).
				Post("/categories").
				Times(1).
				Reply(201).
				JSON(mockCategoryCreate)
			// Mock the request to get and refresh the resource.
			gock.New(testURL).
				Get("/categories/" + mockCategoryCreate.Slug).
				Times(2).
				Reply(200).
				JSON(mockCategory)
			// Mock the request to resolve the version slug from its ID.
			gock.New(testURL).
				Get("/version").
				Persist().
				Reply(200).
				JSON(mockVersionList)
		},
		Check: check,
	}
}

// TestCategoryResource performs basic functionality testing of successfully
// creating, reading, updating, importing, and deleting a category resource.
func TestCategoryResource(t *testing.T) {
	updatedCategory := mockCategory
	updatedCategory.Title = "My Updated Title"

	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the category.
			createCategoryTestStep(resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"created_at",
					mockCategoryCreate.CreatedAt,
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"id",
					mockCategoryCreate.ID,
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"order",
					fmt.Sprintf("%v", mockCategoryCreate.Order),
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"project",
					mockCategoryCreate.Project,
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"reference",
					fmt.Sprintf("%v", mockCategoryCreate.Reference),
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"slug",
					mockCategoryCreate.Slug,
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"title",
					mockCategoryCreate.Title,
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"type",
					mockCategoryCreate.Type,
				),
				resource.TestCheckResourceAttr(
					"readme_category.test",
					"version_id",
					mockCategoryCreate.Version.ID,
				),
			)),
			// Test updating the category.
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "My Updated Title"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				PreConfig: func() {
					// Ensure any existing mocks are removed.
					gock.OffAll()
					// Read current category.
					gock.New(testURL).
						Get("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(200).
						JSON(mockCategory)
					// Update resource.
					gock.New(testURL).
						Put("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(200).
						JSON(updatedCategory)
					// Post-update read.
					gock.New(testURL).
						Get("/categories/" + mockCategory.Slug).
						Times(2).
						Reply(200).
						JSON(updatedCategory)
					// Post-test delete
					gock.New(testURL).
						Delete("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(204)
					// request to lookup version slug from id.
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_category.test",
						"title",
						"My Updated Title",
					),
				),
			},
			// Test importing.
			{
				ResourceName:      "readme_category.test",
				ImportState:       true,
				ImportStateId:     mockCategory.Slug,
				ImportStateVerify: true,
				PreConfig: func() {
					// Ensure any existing mocks are removed.
					gock.OffAll()
					gock.New(testURL).
						Get("/categories/" + mockCategory.Slug).
						Times(2).
						Reply(200).
						JSON(updatedCategory)
					gock.New(testURL).
						Delete("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(204)
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
					gock.New(testURL).
						Get("/categories").
						MatchParam("perPage", "100").
						MatchParam("page", "1").
						Persist().
						Reply(200).
						AddHeader("link", `'<>; rel="next", <>; rel="prev", <>; rel="last"'`).
						AddHeader("x-total-count", "1").
						JSON(mockCategoryList)
				},
			},
		},
	})
}

// TestCategoryResource_Validation_Error tests that an error is returned when
// the category 'type' attribute is invalid. This tests attribute validation.
func TestCategoryResource_Validation_Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + `"
					type  = "invalid"
				}`,
				ExpectError: regexp.MustCompile(
					"Category type must be 'guide' or 'reference'",
				),
			},
		},
	})
}

// TestCategoryResource_Create_Error tests that an error is returned when the
// API responds in error upon creation.
func TestCategoryResource_Create_Error(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + `"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				ExpectError: regexp.MustCompile("Unable to create category"),
				PreConfig: func() {
					// Create resource.
					gock.New(testURL).
						Post("/categories").
						Times(1).
						Reply(400).
						JSON(mockAPIError)
				},
			},
		},
	})
}

// TestCategoryResource_Post_Create_Read_Error tests that an error is returned
// when the API responds in error when reading a newly created resource before
// writing its state.
func TestCategoryResource_Post_Create_Read_Error(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + `"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				ExpectError: regexp.MustCompile("Unable to create category"),
				PreConfig: func() {
					gock.OffAll()
					// Create resource successfully.
					gock.New(testURL).
						Post("/categories").
						Times(1).
						Reply(201).
						JSON(mockCategoryCreate)
					// Fail in post-create read.
					gock.New(testURL).
						Get("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(400).
						JSON(mockAPIError)
				},
			},
		},
	})
}

// TestCategoryResource_Read_Error tests that a read error is returned when the
// API responds with a 404 when requesting an existing category (e.g. on an
// update).
func TestCategoryResource_Read_Error(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the category.
			createCategoryTestStep(nil),
			// Test when response is 404.
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + `_update"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				ExpectError: regexp.MustCompile("Unable to read category"),
				PreConfig: func() {
					gock.OffAll()
					// Return a 404 on a read request on an existing category.
					gock.New(testURL).
						Get("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(404).
						JSON(mockAPIError)
					gock.New(testURL).
						Delete("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(204)
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
				},
			},
		},
	})
}

// TestCategoryResource_Update_Error tests that a error is returned when the
// API responds with a 400 when attempting to update an existing category.
func TestCategoryResource_Update_Error(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the category.
			createCategoryTestStep(nil),
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + `_update"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				ExpectError: regexp.MustCompile("Unable to update category"),
				PreConfig: func() {
					gock.OffAll()
					// Request existing category.
					gock.New(testURL).
						Get("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(200).
						JSON(mockCategory)
					// Return a 400 on update (PUT).
					gock.New(testURL).
						Put("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(400).
						JSON(mockAPIError)
					// Post-test delete.
					gock.New(testURL).
						Delete("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(204)
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
				},
			},
		},
	})
}

// TestCategoryResource_Post_Update_Read_Error tests that an error is returned
// when the API responds in error when reading a newly created resource before
// writing its state.
func TestCategoryResource_Post_Update_Read_Error(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the category.
			createCategoryTestStep(nil),
			{
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + ` update"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				ExpectError: regexp.MustCompile("Unable to update category"),
				PreConfig: func() {
					gock.OffAll()
					// Request existing category.
					gock.New(testURL).
						Get("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(200).
						JSON(mockCategory)
					// Update the category.
					gock.New(testURL).
						Put("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(200).
						JSON(mockCategory)
					// Request updated category.
					gock.New(testURL).
						Get("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(400).
						JSON(mockAPIError)
					// Post-test delete.
					gock.New(testURL).
						Delete("/categories/" + mockCategory.Slug).
						Times(1).
						Reply(204)
					// Lookup version.
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
				},
			},
		},
	})
}

// TestCategoryResource_Delete_Error tests that an error is returned when the
// API responds with an error upon deleting a category.
func TestCategoryResource_Delete_Error(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the category.
			createCategoryTestStep(nil),
			{
				// This is a destroy test.
				Destroy:     true,
				ExpectError: regexp.MustCompile("Unable to delete category"),
				Config: providerConfig + `resource "readme_category" "test" {
					title = "` + mockCategoryCreate.Title + `"
					type  = "` + mockCategoryCreate.Type + `"
				}`,
				PreConfig: func() {
					gock.OffAll()
					// Request existing category.
					gock.New(testURL).
						Get("/categories/" + mockCategoryCreate.Slug).
						Times(3).
						Reply(200).
						JSON(mockCategory)
					// Test failed delete, expect error.
					gock.New(testURL).
						Delete("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(404).
						JSON(mockAPIError)
					// Post-test delete (successful).
					gock.New(testURL).
						Delete("/categories/" + mockCategoryCreate.Slug).
						Times(1).
						Reply(204)
					// Lookup version.
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
				},
			},
		},
	})
}
