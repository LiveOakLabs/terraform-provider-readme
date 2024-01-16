package readme

import "github.com/liveoaklabs/readme-api-go-client/readme"

var mockChangelogs = []readme.Changelog{
	{
		ID:        "1",
		Title:     "Test Changelog",
		Type:      "added",
		Slug:      "test-changelog",
		Body:      "This is a test changelog.",
		CreatedAt: "2021-01-01T00:00:00.000Z",
		UpdatedAt: "2021-01-01T00:00:00.000Z",
		Hidden:    true,
		HTML:      "<p>This is a test changelog.</p>",
		Revision:  1,
		Metadata: readme.DocMetadata{
			Image:       []any{"https://example.com/image.png"},
			Description: "This is a test changelog.",
			Title:       "Test Changelog",
		},
		Algolia: readme.DocAlgolia{
			UpdatedAt: "2021-01-01T00:00:00.000Z",
		},
	},
	{
		ID:        "2",
		Title:     "Test Changelog 2",
		Type:      "added",
		Slug:      "test-changelog-2",
		Body:      "This is a test changelog 2.",
		CreatedAt: "2021-01-01T00:00:00.000Z",
		UpdatedAt: "2021-01-01T00:00:00.000Z",
		Hidden:    true,
		HTML:      "<p>This is a test changelog 2.</p>",
		Revision:  1,
		Metadata: readme.DocMetadata{
			Image:       []any{"https://example.com/image.png"},
			Description: "This is a test changelog 2.",
			Title:       "Test Changelog 2",
		},
		Algolia: readme.DocAlgolia{
			UpdatedAt: "2021-01-01T00:00:00.000Z",
		},
	},
}
