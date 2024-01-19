// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"regexp"
	"strings"
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
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(3).Reply(200).JSON(mockDoc)
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
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(3).Reply(200).JSON(mockDoc)

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
					This is a document.
				`,
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
					This is a document.
				`, mockDoc.Title),
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
					This is a document.
				`, mockDoc.Type,
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
			expect.Body = strings.Trim(expect.Body, "\n")

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
							escapeNewlines(expect.Body),
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

// Test when the 'user' value changes between the apply and post-apply refresh.
func TestDocResource_User_Attribute_Changes(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	expectedDoc := mockDoc

	updatedDoc := mockDoc
	updatedDoc.User = "updated-user"

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Body, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					docCommonGocks()
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"user",
						expectedDoc.User,
					),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "updated body"
						category = "%s"
						type     = "%s"
					}`,
					updatedDoc.Title, updatedDoc.Category, updatedDoc.Type,
				),
				PreConfig: func() {
					updatedDoc.Body = "updated body"
					docCommonGocks()
					// First request responds with the original user.
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Put("/docs").Times(1).Reply(200).JSON(updatedDoc)
					// Post-update request has the updated user.
					gock.New(testURL).Get("/docs/" + updatedDoc.Slug).Times(3).Reply(200).JSON(updatedDoc)
					gock.New(testURL).Delete("/docs/" + updatedDoc.Slug).Times(1).Reply(204)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"user",
						updatedDoc.User,
					),
				),
			},
		},
	})
}

// Test changing the 'hidden' attribute in the doc resource (not front matter).
func TestDocResource_Hidden_Attribute_Changes(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	expectedDoc := mockDoc

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create doc with hidden initially unset.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "body"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = "body"
					expectedDoc.Hidden = false // ReadMe API default.

					docCommonGocks()
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"hidden",
						"false",
					),
				),
			},
			// Change the 'hidden' attribute to true.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "body"
						category = "%s"
						hidden   = true
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = "body"
					expectedDoc.Hidden = true

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"hidden",
						"true",
					),
				),
			},
			// Change to setting the 'hidden' attribute in front matter.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
					    body     = "---\nhidden: false\n---\nbody"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = "---\nhidden: false\n---\nbody"
					expectedDoc.Hidden = false

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(4).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"hidden",
						"false",
					),
				),
			},
			// Change back to setting 'hidden' on the resource.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "body"
						category = "%s"
						hidden   = true
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = "body"
					expectedDoc.Hidden = true

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"hidden",
						"true",
					),
				),
			},
			// Change 'hidden' to false.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "body"
						category = "%s"
						hidden   = false
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = "body"
					expectedDoc.Hidden = false

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"hidden",
						"false",
					),
				),
			},
			// Remove the 'hidden' attribute. It remains unchanged.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "body"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = "body"
					expectedDoc.Hidden = false

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)

					// Test cleanup.
					gock.New(testURL).Delete("/docs/" + expectedDoc.Slug).Times(1).Reply(204)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"hidden",
						"false",
					),
				),
			},
		},
	})
}

// Test changing the 'order' attribute in the doc resource (not front matter).
func TestDocResource_Order_Attribute_Changes(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	expectedDoc := mockDoc

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create doc with order set.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						order    = 1
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Body, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					expectedDoc.Order = 1

					docCommonGocks()
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"1",
					),
				),
			},
			// Change the order attribute.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						order    = 2
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Body, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Order = 2

					docCommonGocks()
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"2",
					),
				),
			},
			// Change to setting order in front matter.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "---\norder: 3\n---\nbody"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					expectedDoc.Body = "---\norder: 3\n---\nbody"
					expectedDoc.Order = 3

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(4).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"3",
					),
				),
			},
			// Change back to setting order in the resource.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
					    body     = "%s"
						category = "%s"
						order    = 4
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Body, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Body = mockDoc.Body
					expectedDoc.Order = 4

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(4).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"4",
					),
				),
			},
			// Remove the order attribute. It is left unchanged.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "%s"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Body, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					expectedDoc.Order = 4

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)

					// Post-test deletion
					gock.New(testURL).Delete("/docs/" + expectedDoc.Slug).Times(1).Reply(204)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"4",
					),
				),
			},
		},
	})
}

// Test changing the 'order' attribute in the doc front matter.
func TestDocResource_Order_FrontMatter_Changes(t *testing.T) {
	// Close all gocks after completion.
	defer gock.OffAll()

	expectedDoc := mockDoc
	expectedDoc.Body = "---\norder: 1\n---\nbody"
	expectedDoc.Order = 1

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create doc with order set in front matter.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "---\norder: 1\n---\nbody"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					docCommonGocks()
					gock.New(testURL).Post("/docs").Times(1).Reply(201).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(3).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"1",
					),
				),
			},
			// Change the order attribute in front matter.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "---\norder: 2\n---\nbody"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					expectedDoc.Body = "---\norder: 2\n---\nbody"
					expectedDoc.Order = 2

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(4).Reply(200).JSON(expectedDoc)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"2",
					),
				),
			},
			// Remove the order attribute.
			{
				Config: providerConfig + fmt.Sprintf(`
					resource "readme_doc" "test" {
						title    = "%s"
						body     = "body"
						category = "%s"
						type     = "%s"
					}`,
					expectedDoc.Title, expectedDoc.Category, expectedDoc.Type,
				),
				PreConfig: func() {
					gock.OffAll()
					expectedDoc.Body = `body`
					expectedDoc.Order = 999

					docCommonGocks()
					gock.New(testURL).Put("/docs/" + expectedDoc.Slug).Times(1).Reply(200).JSON(expectedDoc)
					gock.New(testURL).Get("/docs/" + expectedDoc.Slug).Times(4).Reply(200).JSON(expectedDoc)

					// Post-test deletion
					gock.New(testURL).Delete("/docs/" + expectedDoc.Slug).Times(1).Reply(204)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_doc.test",
						"order",
						"999",
					),
				),
			},
		},
	})
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
					gock.New(testURL).Get("/docs/" + mockDoc.Slug).Times(3).Reply(200).JSON(mockDoc)
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
					gock.New(testURL).Get("/docs/" + "new-slug").Times(4).Reply(200).JSON(renamed)

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
