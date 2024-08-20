package readme

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestDocDataSource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					gock.New(testURL).
						Get("/docs").
						Path(mockDoc.Slug).
						Persist().
						Reply(200).
						JSON(mockDoc)
				},
				Config: testProviderConfig + fmt.Sprintf(`
					data "readme_doc" "test" {
						slug = "%s"
					}`,
					mockDoc.Slug,
				),
				Check: docResourceCommonChecks(mockDoc, "data."),
			},
		},
	})
}
