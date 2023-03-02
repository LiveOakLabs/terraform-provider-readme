package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestImageResource(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	// mockImageResponse represents the response from the API when uploading an image.
	// It's a slice of interface{} because the API returns a JSON array of mixed types.
	mockImageResponse := []any{
		"https://files.readme.io/c6f07db-example.png",
		"example.png",
		1,
		1,
		"#000000",
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "readme_image" "test" {
					source = "../examples/resources/readme_image/example.png"
				}`,
				PreConfig: func() {
					// Upload the image.
					gock.New("https://dash.readme.com/api/images").
						Post("/image-upload").
						Times(1).
						Reply(200).
						JSON(mockImageResponse)
					// Post-create read and refresh.
					gock.New(mockImageResponse[0].(string)).
						Head("/").
						Times(1).
						Reply(200)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_image.test",
						"url",
						mockImageResponse[0].(string),
					),
					resource.TestCheckResourceAttr(
						"readme_image.test",
						"filename",
						mockImageResponse[1].(string),
					),
					resource.TestCheckResourceAttr(
						"readme_image.test",
						"width",
						fmt.Sprintf("%v", mockImageResponse[2].(int)),
					),
					resource.TestCheckResourceAttr(
						"readme_image.test",
						"height",
						fmt.Sprintf("%v", mockImageResponse[3].(int)),
					),
					resource.TestCheckResourceAttr(
						"readme_image.test",
						"color",
						mockImageResponse[4].(string),
					),
				),
			},
		},
	})
}

func TestImageResource_Errors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "readme_image" "test" {
					source = "invalid/path/to/image.png"
				}`,
				ExpectError: regexp.MustCompile("Unable to read source image file"),
			},
		},
	})
}
