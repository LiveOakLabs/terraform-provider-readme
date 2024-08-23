// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestChangelogResource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()

	mockUpdatedChangelog := mockChangelogs[0]
	mockUpdatedChangelog.Title = "Updated Title"
	mockUpdatedChangelog.Body = fmt.Sprintf(
		"---\ntitle: %s\n---\n",
		mockUpdatedChangelog.Title,
	) + mockUpdatedChangelog.Body

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Persist().
						Reply(200).
						JSON(mockChangelogs[0])
					gock.New(testURL).
						Post("/changelogs").
						Persist().
						Reply(201).
						JSON(mockChangelogs[0])
				},
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						title = "` + mockChangelogs[0].Title + `"
						type  = "` + mockChangelogs[0].Type + `"
						body  = "` + mockChangelogs[0].Body + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"id",
						mockChangelogs[0].ID,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"title",
						mockChangelogs[0].Title,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"type",
						mockChangelogs[0].Type,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"slug",
						mockChangelogs[0].Slug,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"body",
						mockChangelogs[0].Body,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"created_at",
						mockChangelogs[0].CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"updated_at",
						mockChangelogs[0].UpdatedAt,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"revision",
						fmt.Sprintf("%d", mockChangelogs[0].Revision),
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"hidden",
						fmt.Sprintf("%t", mockChangelogs[0].Hidden),
					),
				),
			},
			// Test updating.
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Put("/changelogs").
						Times(1).
						Reply(200).
						JSON(mockUpdatedChangelog)
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(3).
						Reply(200).
						JSON(mockUpdatedChangelog)
					gock.New(testURL).
						Delete("/changelogs/" + mockChangelogs[0].Slug).
						Times(1).
						Reply(204)
				},
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						title = "` + mockUpdatedChangelog.Title + `"
						type  = "` + mockUpdatedChangelog.Type + `"
						body  = "` + escapeNewlines(mockUpdatedChangelog.Body) + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"title",
						mockUpdatedChangelog.Title,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"type",
						mockUpdatedChangelog.Type,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"body",
						mockUpdatedChangelog.Body,
					),
				),
			},
			// Test updating with front matter.
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Put("/changelogs").
						Times(1).
						Reply(200).
						JSON(mockUpdatedChangelog)
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(1).
						Reply(200).
						JSON(mockChangelogs[0])
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(2).
						Reply(200).
						JSON(mockUpdatedChangelog)
					gock.New(testURL).
						Delete("/changelogs/" + mockChangelogs[0].Slug).
						Times(1).
						Reply(204)
				},
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						body  = "` + escapeNewlines(mockUpdatedChangelog.Body) + `"
						type  = "` + mockUpdatedChangelog.Type + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"title",
						mockUpdatedChangelog.Title,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"body",
						mockUpdatedChangelog.Body,
					),
				),
			},
			// Test import.
			{
				ResourceName:  "readme_changelog.test",
				ImportState:   true,
				ImportStateId: mockChangelogs[0].Slug,
				PreConfig: func() {
					// Ensure any existing mocks are removed.
					gock.OffAll()
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(2).
						Reply(200).
						JSON(mockChangelogs[0])
				},
			},
			// Test updating with no title results in error.
			{
				ExpectError: regexp.MustCompile("Title is not set"),
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						body  = "no title is set with front matter or attribute"
				}`,
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(1).
						Reply(200).
						JSON(mockChangelogs[0])
					gock.New(testURL).Delete("/changelogs/" + mockChangelogs[0].Slug).Times(1).Reply(204)
				},
			},
		},
	})
}

func TestChangelogResourceFrontMatter(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()

	mockChangelog := mockChangelogs[0]
	mockChangelog.Title = "turtle"
	mockChangelog.Body = "---\ntitle: turtle\n---\nturtles are cool"

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Get("/changelogs/" + mockChangelog.Slug).
						Persist().
						Reply(200).
						JSON(mockChangelog)
					gock.New(testURL).
						Post("/changelogs").
						Persist().
						Reply(201).
						JSON(mockChangelog)
					gock.New(testURL).
						Delete("/changelogs/" + mockChangelog.Slug).
						Times(1).
						Reply(204)
				},
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						body  = "` + escapeNewlines(mockChangelog.Body) + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"id",
						mockChangelog.ID,
					),
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"title",
						"turtle",
					),
				),
			},
		},
	})
}

// Test that a changelog gets re-created when it's deleted externally.
func TestChangelogResource_ReCreate(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Persist().
						Reply(200).
						JSON(mockChangelogs[0])
					gock.New(testURL).
						Post("/changelogs").
						Persist().
						Reply(201).
						JSON(mockChangelogs[0])
				},
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						title = "` + mockChangelogs[0].Title + `"
						type  = "` + mockChangelogs[0].Type + `"
						body  = "` + mockChangelogs[0].Body + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"id",
						mockChangelogs[0].ID,
					),
				),
			},
			{
				PreConfig: func() {
					gock.OffAll()
					// Initial read returns 404.
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(1).
						Reply(404).
						JSON(map[string]string{"message": "Not Found"})
					// Create a new changelog.
					gock.New(testURL).
						Post("/changelogs").
						Persist().
						Reply(201).
						JSON(mockChangelogs[0])
					// Post-create read.
					gock.New(testURL).
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Times(2).
						Reply(200).
						JSON(mockChangelogs[0])
					// Post-run refresh.
					gock.New(testURL).
						Delete("/changelogs/" + mockChangelogs[0].Slug).
						Times(1).
						Reply(204)
				},
				Config: testProviderConfig + `
					resource "readme_changelog" "test" {
						title = "` + mockChangelogs[0].Title + `"
						type  = "` + mockChangelogs[0].Type + `"
						body  = "` + mockChangelogs[0].Body + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_changelog.test",
						"id",
						mockChangelogs[0].ID,
					),
				),
			},
		},
	})
}
