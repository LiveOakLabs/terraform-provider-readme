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
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_api_specification" "test" {
						definition = "%s"
					}`,
					testdata.APISpecificationDefinitionSrc,
				),
				PreConfig: testdata.APISpecificationCreateRespond(mockVersionList),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("readme_api_specification.test", "id", testdata.APISpecifications[0].ID),
					resource.TestCheckResourceAttr("readme_api_specification.test", "last_synced", testdata.APISpecifications[0].LastSynced),
					resource.TestCheckResourceAttr("readme_api_specification.test", "source", testdata.APISpecifications[0].Source),
					resource.TestCheckResourceAttr("readme_api_specification.test", "title", testdata.APISpecifications[0].Title),
					resource.TestCheckResourceAttr("readme_api_specification.test", "type", testdata.APISpecifications[0].Type),
					resource.TestCheckResourceAttr("readme_api_specification.test", "version", testdata.APISpecifications[0].Version),
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
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_api_specification" "test" {
						definition      = "%s"
						delete_category = true
					}`,
					testdata.APISpecificationDefinitionSrc,
				),
				PreConfig: testdata.APISpecificationDeleteCategoryRespond(mockVersionList),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("readme_api_specification.test", "id", testdata.APISpecifications[0].ID),
				),
			},
		},
	})
}
