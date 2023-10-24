// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

func TestDocResource(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test successful creation.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Title, mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					docCommonGocks()
					// Mock the request to create the resource.
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(mockDoc)
					// Mock the request to get and refresh the resource.
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(2).Reply(200).JSON(mockDoc)
				},
				Check: docResourceCommonChecks(mockDoc, ""),
			},

			// Test that a new doc gets created if it's deleted outside of Terraform.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Title, mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					// Pre-update read.
					mockAPIError.Error = "DOC_NOTFOUND"
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(1).Reply(404).JSON(mockAPIError)

					gock.New(testURL).
						Post("/docs").
						Path("search").
						Times(1).
						Reply(200).
						JSON(readme.DocSearchResult{})

					// Mock the request to create the resource.
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(mockDoc)
					// Mock the request to get and refresh the resource.
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(2).Reply(200).JSON(mockDoc)

					gock.New(testURL).Delete("/docs/" + mockDoc.Slug).Times(1).Reply(204)
				},
				Check: docResourceCommonChecks(mockDoc, ""),
			},

			// Test update results in error when the pre-update read fails.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "Update"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(1).Reply(400).JSON(mockAPIError)
				},
				ExpectError: regexp.MustCompile("DOC_NOTFOUND"),
			},

			// Test update results in error when the update action fails.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "Update 2"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					gock.New(testURL).
						Get("/docs/" + mockDoc.Slug).
						Persist().
						Reply(200).
						JSON(mockDoc)
					gock.New(testURL).
						Put("/docs/" + mockDoc.Slug).
						Persist().
						Reply(400).
						JSON(mockAPIError)
					gock.New(testURL).Post("/docs/").Persist().Reply(400).JSON(mockAPIError)
				},
				ExpectError: regexp.MustCompile("Unable to update doc"),
			},

			// Test update results in error when post-update read fails.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "Update"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(1).Reply(200).JSON(mockDoc)
					// Mock the request to create the resource.
					mockDoc.Title = "Update"
					gock.New(testURL).Put("/docs/" + mockDoc.Slug).Times(1).Reply(200).JSON(mockDoc)
					gock.New(testURL).
						Get("/docs/" + mockDoc.Slug).
						Times(2).
						Reply(400).
						JSON(mockAPIError)
				},
				ExpectError: regexp.MustCompile("Unable to update doc"),
			},

			// Test delete error when API responds with 400.
			{
				Destroy: true,
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Title, mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(2).Reply(200).JSON(mockDoc)
					gock.New(testURL).Delete("/docs/" + mockDoc.Slug).Times(1).Reply(400)
				},
				ExpectError: regexp.MustCompile("Unable to delete doc"),
			},

			// Test successful update when title changes.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "Update"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					// Pre-update read.
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(1).Reply(200).JSON(mockDoc)

					// Update responses chain.
					mockDoc.Title = "Update"
					gock.New(testURL).Put("/docs/" + mockDoc.Slug).Times(1).Reply(200).JSON(mockDoc)
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(2).Reply(200).JSON(mockDoc)
					gock.New(testURL).Delete("/docs/" + mockDoc.Slug).Times(1).Reply(204)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"title",
						"Update",
					),
				),
			},

			// Test import.
			{
				ResourceName:  "readme_doc.test",
				ImportState:   true,
				ImportStateId: mockDoc.Slug,
				PreConfig: func() {
					// Ensure any existing mocks are removed.
					gock.OffAll()
					docCommonGocks()
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(2).Reply(200).JSON(mockDoc)
					gock.New(testURL).Delete("/docs/" + mockDoc.Slug).Times(1).Reply(204)
				},
			},
		},
	})
}

func TestDocResource_Create_Errors(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	testCases := []struct {
		desc           string
		PostCode       int
		PostResponse   any
		GetCode        int
		GetResponse    any
		DeleteCode     int
		DeleteResponse any
	}{
		{
			desc:         "it returns an error when API responds with 400 when POSTing a doc",
			PostCode:     400,
			PostResponse: mockAPIError,
		},
		{
			desc:         "it returns an error when API responds with 400 when GETing the created doc",
			PostCode:     201,
			PostResponse: mockDoc,
			GetCode:      400,
			GetResponse:  mockAPIError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				IsUnitTest:               true,
				ProtoV6ProviderFactories: testProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: providerConfig + fmt.Sprintf(`
							resource "readme_doc" "test" {
								title    = "Update"
								category = "%s"
								type     = "%s"
							}`,
							mockDoc.Category, mockDoc.Type,
						),
						PreConfig: func() {
							if testCase.PostCode != 0 {
								// Mock the request to create the resource.
								gock.New(testURL).Post("/docs").
									Times(1).
									Reply(testCase.PostCode).
									JSON(testCase.PostResponse)
							}
							if testCase.GetCode != 0 {
								// Mock the request to get and refresh the resource.
								gock.New(testURL).
									Get("/docs/" + mockDoc.Slug).
									Times(1).
									Reply(testCase.PostCode).
									JSON(testCase.GetResponse)
							}
							if testCase.DeleteCode != 0 {
								gock.New(testURL).
									Delete("/docs/" + mockDoc.Slug).
									Times(1).
									Reply(204)
							}
						},
						ExpectError: regexp.MustCompile(`Unable to create doc`),
					},
				},
			})
		})
	}
}

// TestDocResource_FrontMatter is a series of tests for use of front matter in a body with resource attributes.
// Attribute values should always take precedence over front matter, but front matter should be used as a value if the
// corresponding attribute isn't otherwise specified.
func TestDocResource_FrontMatter(t *testing.T) {
	testCases := []struct {
		attributes string
		desc       string
		expect     readme.DocParams
	}{
		{
			desc: "title value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Body: `
					---
					title: ignored
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "title value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Body: fmt.Sprintf(`
					---
					title: %s
					---
					This is a document.`,
					mockDoc.Title,
				),
			},
			attributes: fmt.Sprintf(
				`
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "category value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Body: `
					---
					category: ignored
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "category value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Body: fmt.Sprintf(`
					---
					category: %s
					---
					This is a document.`,
					mockDoc.Category,
				),
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
				`,
				mockDoc.Title, mockDoc.Type,
			),
		},
		{
			desc: "category_slug value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Body: `
					---
					categorySlug: ignored
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category_slug = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockCategory.Slug,
			),
		},
		{
			desc: "category_slug value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Body: fmt.Sprintf(`
					---
					categorySlug: %s
					---
					This is a document.`, mockCategory.Slug,
				),
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
				`,
				mockDoc.Title, mockDoc.Type,
			),
		},
		{
			desc: "hidden value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Hidden: boolPoint(true), // Ensure default override.
				Body: `
					---
					hidden: false
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
					hidden   = true
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "hidden value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Hidden: boolPoint(false), // Ensure default override.
				Body: `
					---
					hidden: false
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "order value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Order: intPoint(mockDoc.Order), // Ensure default override.
				Body: `
					---
					order: 99
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
					order    = %v
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category, mockDoc.Order,
			),
		},
		{
			desc: "order value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Order: intPoint(99), // Ensure default override.
				Body: `
					---
					order: 99
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "parent_doc value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Body: `
					---
					parentDoc: ignored
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
					parent_doc = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category, mockDoc.ParentDoc,
			),
		},
		{
			desc: "parent_doc value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Body: fmt.Sprintf(`
					---
					parentDoc: '%s'
					---
					This is a document.`, mockDoc.ParentDoc),
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "type value is set from attribute and not front matter when attribute is set",
			expect: readme.DocParams{
				Body: `
					---
					type: ignored
					---
					This is a document.`,
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					type     = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Type, mockDoc.Category,
			),
		},
		{
			desc: "type value is set from front matter when attribute is not set",
			expect: readme.DocParams{
				Body: fmt.Sprintf(`
					---
					type: %s
					---
					This is a document.`, mockDoc.Type,
				),
			},
			attributes: fmt.Sprintf(
				`
					title    = "%s"
					category = "%s"
				`,
				mockDoc.Title, mockDoc.Category,
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			// Set up the doc we're testing, derived from the common mock with the expected attributes overridden.
			expect := mockDoc
			expect.Body = removeIndents(testCase.expect.Body)

			if testCase.expect.Hidden != nil {
				expect.Hidden = *testCase.expect.Hidden
			}
			if testCase.expect.Order != nil {
				expect.Order = *testCase.expect.Order
			}

			// Close all gocks after the test completes.
			defer gock.OffAll()

			resource.Test(t, resource.TestCase{
				IsUnitTest:               true,
				ProtoV6ProviderFactories: testProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: providerConfig + fmt.Sprintf(
							`
								resource "readme_doc" "test" {
										body = chomp("%s")
										%s
								}
							`,
							replaceNewlines(expect.Body),
							testCase.attributes,
						),
						PreConfig: func() {
							gock.OffAll()
							docCommonGocks()
							// Mock the request to create the resource.
							gock.New(testURL).Post("/docs").Times(3).Reply(201).JSON(expect)
							// Mock the request to get and refresh the resource.
							gock.New(testURL).
								Get("/docs/" + mockDoc.Slug).
								Persist().
								Reply(200).
								JSON(expect)
							// Mock the post-test delete.
							gock.New(testURL).Delete("/docs/" + expect.Slug).Times(1).Reply(204)
						},
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"category",
								expect.Category,
							),
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"category_slug",
								mockCategory.Slug,
							),
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"hidden",
								fmt.Sprintf("%v", expect.Hidden),
							),
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"order",
								fmt.Sprintf("%v", expect.Order),
							),
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"parent_doc",
								expect.ParentDoc,
							),
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"parent_doc_slug",
								mockDocParent.Slug,
							),
							resource.TestCheckResourceAttr(
								"readme_doc.test",
								"title",
								expect.Title,
							),
							resource.TestCheckResourceAttr("readme_doc.test", "type", expect.Type),
						),
					},
				},
			})
		})
	}
}

// TestDocRenamedSlugResource tests that a doc can be created and continue to
// be managed by Terraform when the slug is changed outside of Terraform by
// using the "get_slug" attribute.
func TestDocRenamedSlugResource(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	renamed := mockDoc
	renamed.Slug = "new-slug"
	renamedSearch := mockDocSearchResponse
	renamedSearch.Results[0].Slug = "new-slug"

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test successful creation.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					mockDoc.Title, mockDoc.Body, mockDoc.Category, mockDoc.Type,
				),
				PreConfig: func() {
					docCommonGocks()
					// Mock the request to create the resource.
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(mockDoc)
					// Mock the request to get and refresh the resource.
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(2).Reply(200).JSON(mockDoc)
				},
				Check: docResourceCommonChecks(mockDoc, ""),
			},

			// Test that the doc can be renamed outside of Terraform and
			// continue to be managed by Terraform.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					renamed.Title, renamed.Body, renamed.Category, renamed.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()

					// Original slug is not found.
					docNotFoundAPIError := mockAPIError
					docNotFoundAPIError.Error = "DOC_NOTFOUND"
					docNotFoundAPIError.Message = "Doc not found"
					gock.New(testURL).Get("/docs/" + "a-test-doc").Times(1).Reply(404).JSON(docNotFoundAPIError)

					// The slug won't exist, so the provider does a search by ID.
					gock.New(testURL).
						Post("/docs").
						Path("search").
						Times(1).
						Reply(200).
						JSON(mockDocSearchResponse)

					// The matched doc is requested from the search results.
					// It's also requested again after the rename.
					gock.New(testURL).Get("/docs/" + "new-slug").Times(3).Reply(200).JSON(renamed)

					// An update is triggered to match state with the new slug.
					gock.New(testURL).Put("/docs/" + "new-slug").Times(1).Reply(200).JSON(renamed)

					// Post-test deletion
					gock.New(testURL).Delete("/docs/" + "new-slug").Times(1).Reply(204)
				},
				Check: docResourceCommonChecks(renamed, ""),
			},
		},
	})
}
