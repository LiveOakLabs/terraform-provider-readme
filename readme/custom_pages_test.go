package readme

import "github.com/liveoaklabs/readme-api-go-client/readme"

// mockCustomPages represents a ReadMe response for a list of custom pages.
// This is used throughout the custom pages tests.
var mockCustomPages = []readme.CustomPage{
	{
		Algolia: readme.DocAlgolia{
			PublishPending: false,
			RecordCount:    0,
			UpdatedAt:      "2022-10-06T20:25:15.000Z",
		},
		Body:       "This is a test custom page.",
		CreatedAt:  "2022-10-04T20:25:15.000Z",
		Fullscreen: false,
		HTML:       "",
		Hidden:     true,
		HTMLMode:   false,
		ID:         "5f7b5f5b1b9b4c001b8b4b1a",
		Metadata: readme.DocMetadata{
			Image: []any{
				"https://example.com/image.png",
				"image.png",
				512,
				512,
				"#000000",
			},
			Title:       "",
			Description: "",
		},
		Revision:  4,
		Slug:      "test-custom-page",
		Title:     "Test Custom Page",
		UpdatedAt: "2023-04-12T20:25:15.000Z",
	},
	{
		Algolia: readme.DocAlgolia{
			PublishPending: false,
			RecordCount:    0,
			UpdatedAt:      "2022-10-06T20:25:15.000Z",
		},
		Body:       "",
		CreatedAt:  "2023-04-06T10:45:39.532Z",
		Fullscreen: false,
		HTML:       "<div>This is a test page with HTML.</div>",
		Hidden:     true,
		HTMLMode:   true,
		ID:         "605c8f5f1b9b4c001b8b4b1b",
		Metadata: readme.DocMetadata{
			Image: []any{
				"https://example.com/image-two.png",
				"image-two.png",
				256,
				256,
				"#ffffff",
			},
			Title:       "",
			Description: "",
		},
		Revision:  1,
		Slug:      "test-custom-page-2",
		Title:     "Test Custom Page 2",
		UpdatedAt: "2023-04-06T10:45:39.532Z",
	},
}
