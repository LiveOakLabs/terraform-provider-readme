package readme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestAPISpecificationDataSource(t *testing.T) {
	defer gock.Off()
	testCases := []struct {
		name   string
		config string
	}{
		{
			name:   "lookup api specification by id",
			config: `data "readme_api_specification" "test" { id = "6398a4a594b26e00885e7ec0" }`,
		},
		{
			name:   "lookup api specification by title",
			config: `data "readme_api_specification" "test" { title = "Test API Spec" }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
								JSON(`[{
									"id": "6398a4a594b26e00885e7ec0",
									"lastSynced": "2022-12-13T16:39:39.512Z",
									"source": "api",
									"title": "Test API Spec",
									"type": "oas",
									"version": "638cf4cfdea3ff0096d1a95a"
								}]
							`)
						},
					},
				},
			})
		})
	}
}
