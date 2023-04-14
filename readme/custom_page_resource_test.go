package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestCustomPageResource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()

	mockUpdatedCustomPage := mockCustomPages[0]
	mockUpdatedCustomPage.Title = "Updated Title"
	mockUpdatedCustomPage.Body = "Updated Body"

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Get("/custompages/" + mockCustomPages[0].Slug).
						Persist().
						Reply(200).
						JSON(mockCustomPages[0])
					gock.New(testURL).
						Post("/custompages").
						Persist().
						Reply(201).
						JSON(mockCustomPages[0])
				},
				Config: providerConfig + `
					resource "readme_custom_page" "test" {
						title = "` + mockCustomPages[0].Title + `"
						body  = "` + mockCustomPages[0].Body + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"id",
						mockCustomPages[0].ID,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"title",
						mockCustomPages[0].Title,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"slug",
						mockCustomPages[0].Slug,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"body",
						mockCustomPages[0].Body,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"created_at",
						mockCustomPages[0].CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"updated_at",
						mockCustomPages[0].UpdatedAt,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"revision",
						fmt.Sprintf("%d", mockCustomPages[0].Revision),
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"fullscreen",
						fmt.Sprintf("%t", mockCustomPages[0].Fullscreen),
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"hidden",
						fmt.Sprintf("%t", mockCustomPages[0].Hidden),
					),
				),
			},
			// Test updating.
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Put("/custompages").
						Times(1).
						Reply(200).
						JSON(mockUpdatedCustomPage)
					gock.New(testURL).
						Get("/custompages/" + mockCustomPages[0].Slug).
						Times(1).
						Reply(200).
						JSON(mockCustomPages[0])
					gock.New(testURL).
						Get("/custompages/" + mockCustomPages[0].Slug).
						Times(1).
						Reply(200).
						JSON(mockUpdatedCustomPage)
					gock.New(testURL).
						Delete("/custompages/" + mockCustomPages[0].Slug).
						Times(1).
						Reply(204)
				},
				Config: providerConfig + `
					resource "readme_custom_page" "test" {
						title = "` + mockUpdatedCustomPage.Title + `"
						body  = "` + mockUpdatedCustomPage.Body + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"title",
						mockUpdatedCustomPage.Title,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"body",
						mockUpdatedCustomPage.Body,
					),
				),
			},
			// Test updating with no title results in error.
			{
				ExpectError: regexp.MustCompile("'title' must be set using the attribute or in the body front matter."),
				Config: providerConfig + `
					resource "readme_custom_page" "test" {
						body  = "no title is set with front matter or attribute"
				}`,
			},
			// Test updating with front matter.
			{
				PreConfig: func() {
					gock.OffAll()
					// Get category list to lookup category.
					gock.New(testURL).
						Put("/custompages").
						Times(1).
						Reply(200).
						JSON(mockUpdatedCustomPage)
					gock.New(testURL).
						Get("/custompages/" + mockCustomPages[0].Slug).
						Times(1).
						Reply(200).
						JSON(mockCustomPages[0])
					gock.New(testURL).
						Get("/custompages/" + mockCustomPages[0].Slug).
						Times(1).
						Reply(200).
						JSON(mockUpdatedCustomPage)
					gock.New(testURL).
						Delete("/custompages/" + mockCustomPages[0].Slug).
						Times(1).
						Reply(204)
				},
				Config: providerConfig + `
					resource "readme_custom_page" "test" {
						body  = "---\ntitle: ` + mockUpdatedCustomPage.Title + `\n---\n` + mockUpdatedCustomPage.Body + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"title",
						mockUpdatedCustomPage.Title,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"body",
						"---\ntitle: "+mockUpdatedCustomPage.Title+"\n---\n"+mockUpdatedCustomPage.Body,
					),
				),
			},
			// Test import.
			{
				ResourceName:  "readme_custom_page.test",
				ImportState:   true,
				ImportStateId: mockCustomPages[0].Slug,
				PreConfig: func() {
					// Ensure any existing mocks are removed.
					gock.OffAll()
					gock.New(testURL).Get("/custompages/" + mockCustomPages[0].Slug).Times(2).Reply(200).JSON(mockCustomPages[0])
					gock.New(testURL).Delete("/custompages/" + mockCustomPages[0].Slug).Times(1).Reply(204)
				},
			},
		},
	})
}

func TestCustomPageResource_HTML(t *testing.T) {
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
						Get("/custompages/" + mockCustomPages[1].Slug).
						Persist().
						Reply(200).
						JSON(mockCustomPages[1])
					gock.New(testURL).
						Post("/custompages").
						Persist().
						Reply(201).
						JSON(mockCustomPages[1])
					gock.New(testURL).
						Delete("/custompages/" + mockCustomPages[1].Slug).
						Times(1).
						Reply(204)
				},
				Config: providerConfig + `
					resource "readme_custom_page" "test" {
						title     = "` + mockCustomPages[1].Title + `"
						html      = "<html><body>` + mockCustomPages[1].HTML + `</body></html>"
						html_mode = true
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"id",
						mockCustomPages[1].ID,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"title",
						mockCustomPages[1].Title,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"slug",
						mockCustomPages[1].Slug,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"body",
						mockCustomPages[1].Body,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"html",
						"<html><body>"+mockCustomPages[1].HTML+"</body></html>",
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"html_clean",
						mockCustomPages[1].HTML,
					),
					resource.TestCheckResourceAttr(
						"readme_custom_page.test",
						"html_mode",
						fmt.Sprintf("%v", mockCustomPages[1].HTMLMode),
					),
				),
			},
		},
	})
}
