// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/terraform-provider-readme/internal/testdata"
	"gopkg.in/h2non/gock.v1"
)

// TestAPISpecificationResource_Create is a happy path test for the Create method.
func TestAPISpecificationResource_Create(t *testing.T) {
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + fmt.Sprintf(`
					resource "readme_api_specification" "test" {
						definition = "%s"
					}`,
					testdata.APISpecificationDefinitionSrc,
				),
				PreConfig: testdata.APISpecificationCreateRespond(mockVersionList),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"id",
						testdata.APISpecifications[0].ID,
					),
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"last_synced",
						testdata.APISpecifications[0].LastSynced,
					),
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"source",
						testdata.APISpecifications[0].Source,
					),
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"title",
						testdata.APISpecifications[0].Title,
					),
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"type",
						testdata.APISpecifications[0].Type,
					),
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"version",
						testdata.APISpecifications[0].Version,
					),
				),
			},
		},
	})
}

// TestAPISpecificationResource_CreateDeleteCategory tests that the
// delete_category flag deletes the category that's created when the API spec
// is created.
func TestAPISpecificationResource_CreateDeleteCategory(t *testing.T) {
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + fmt.Sprintf(`
					resource "readme_api_specification" "test" {
						definition      = "%s"
						delete_category = true
					}`,
					testdata.APISpecificationDefinitionSrc,
				),
				PreConfig: testdata.APISpecificationDeleteCategoryRespond(mockVersionList),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_api_specification.test",
						"id",
						testdata.APISpecifications[0].ID,
					),
				),
			},
		},
	})
}

func TestJsonMatch(t *testing.T) {
	petstoreSpec1 := `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Swagger Petstore",
    "description": "This is a sample server for a pet store.",
    "version": "1.0.0"
  },
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "responses": {
          "200": {
            "description": "A paged array of pets"
          }
        }
      }
    }
  }
}`

	// Same as petstoreSpec1, but the order of keys in the 'info' object is different
	petstoreSpec1DifferentOrder := `
{
  "openapi": "3.0.0",
  "info": {
    "version": "1.0.0",
    "description": "This is a sample server for a pet store.",
    "title": "Swagger Petstore"
  },
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "responses": {
          "200": {
            "description": "A paged array of pets"
          }
        }
      }
    }
  }
}`

	petstoreSpecWithExtra := `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Swagger Petstore",
    "description": "This is a sample server for a pet store.",
    "version": "1.0.0"
  },
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "responses": {
          "200": {
            "description": "A paged array of pets"
          }
        }
      },
      "post": {
        "summary": "Create a pet",
        "operationId": "createPets",
        "responses": {
          "201": {
            "description": "Null response"
          }
        }
      }
    }
  }
}`

	invalidSpec := `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Swagger Petstore",
    "description": "This is a sample server for a pet store.",
    "version": "1.0.0"
  },
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "responses": {
          "200": {
            "description": "A paged array of pets"
          }
        }
      }
    }
  ` // Invalid JSON (missing closing braces)

	tests := []struct {
		name    string
		json1   string
		json2   string
		want    bool
		wantErr bool
	}{
		{
			name:    "Identical OpenAPI specs",
			json1:   petstoreSpec1,
			json2:   petstoreSpec1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Same content, different order",
			json1:   petstoreSpec1,
			json2:   petstoreSpec1DifferentOrder,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Different OpenAPI specs (additional endpoint)",
			json1:   petstoreSpec1,
			json2:   petstoreSpecWithExtra,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid JSON in one spec",
			json1:   petstoreSpec1,
			json2:   invalidSpec,
			want:    false,
			wantErr: true,
		},
	}

	// nolint:varnamelen // Intentional repetition of some values for tests.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonMatch(tt.json1, tt.json2)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonMatch() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("jsonMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
