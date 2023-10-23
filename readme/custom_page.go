package readme

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// customPageDataSourceModel is the data source model used by the
// readme_custom_page and readme_custom_pages data sources.
type customPageDataSourceModel struct {
	Algolia    types.Object `tfsdk:"algolia"`
	Body       types.String `tfsdk:"body"`
	CreatedAt  types.String `tfsdk:"created_at"`
	FullScreen types.Bool   `tfsdk:"fullscreen"`
	HTML       types.String `tfsdk:"html"`
	HTMLMode   types.Bool   `tfsdk:"htmlmode"`
	Hidden     types.Bool   `tfsdk:"hidden"`
	ID         types.String `tfsdk:"id"`
	Metadata   *docMetadata `tfsdk:"metadata"`
	Revision   types.Int64  `tfsdk:"revision"`
	Slug       types.String `tfsdk:"slug"`
	Title      types.String `tfsdk:"title"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// customPageDatasourceMapToModel maps a readme.CustomPage to a customPageDataSourceModel
// for use in the readme_custom_page and readme_custom_pages data sources.
func customPageDatasourceMapToModel(page readme.CustomPage) customPageDataSourceModel {
	image := []string{}
	if page.Metadata.Image != nil {
		for _, i := range page.Metadata.Image {
			image = append(image, fmt.Sprintf("%v", i))
		}
	}

	model := customPageDataSourceModel{
		Title:      types.StringValue(page.Title),
		Slug:       types.StringValue(page.Slug),
		Body:       types.StringValue(page.Body),
		HTML:       types.StringValue(page.HTML),
		HTMLMode:   types.BoolValue(page.HTMLMode),
		FullScreen: types.BoolValue(page.Fullscreen),
		Hidden:     types.BoolValue(page.Hidden),
		Revision:   types.Int64Value(int64(page.Revision)),
		ID:         types.StringValue(page.ID),
		CreatedAt:  types.StringValue(page.CreatedAt),
		UpdatedAt:  types.StringValue(page.UpdatedAt),
		Metadata: &docMetadata{
			Image:       image,
			Title:       page.Metadata.Title,
			Description: page.Metadata.Description,
		},
		Algolia: docModelAlgoliaValue(page.Algolia),
	}

	return model
}

// customPageResourceMapToModel maps a readme.CustomPage to a customPageResourceModel
// for use in the readme_custom_page resource.
func customPageResourceMapToModel(
	page readme.CustomPage,
	plan customPageResourceModel,
) customPageResourceModel {
	if plan.Body.IsUnknown() {
		plan.Body = types.StringValue("")
	}

	if plan.HTML.IsUnknown() {
		plan.HTML = types.StringValue("")
	}

	return customPageResourceModel{
		Algolia:    docModelAlgoliaValue(page.Algolia),
		Body:       plan.Body,
		BodyClean:  types.StringValue(page.Body),
		CreatedAt:  types.StringValue(page.CreatedAt),
		FullScreen: types.BoolValue(page.Fullscreen),
		HTML:       plan.HTML,
		HTMLClean:  types.StringValue(page.HTML),
		HTMLMode:   types.BoolValue(page.HTMLMode),
		Hidden:     types.BoolValue(page.Hidden),
		ID:         types.StringValue(page.ID),
		Metadata:   docModelMetadataValue(page.Metadata),
		Revision:   types.Int64Value(int64(page.Revision)),
		Slug:       types.StringValue(page.Slug),
		Title:      types.StringValue(page.Title),
		UpdatedAt:  types.StringValue(page.UpdatedAt),
	}
}

// customPageDataSourceSchema returns the schema for the
// readme_custom_page and readme_custom_pages data sources.
func customPageDataSourceSchema() map[string]schema.Attribute {
	// nolint:goconst // Attribute descriptions are
	// repeated across resources and data sources.
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			Description: "The title of the custom page.",
			Computed:    true,
		},
		"slug": schema.StringAttribute{
			Description: "The slug of the custom page.",
			Computed:    true,
		},
		"body": schema.StringAttribute{
			Description: "The body of the custom page.",
			Computed:    true,
		},
		"html": schema.StringAttribute{
			Description: "The HTML of the custom page.",
			Computed:    true,
		},
		"htmlmode": schema.BoolAttribute{
			Description: "Whether the custom page is in HTML mode.",
			Computed:    true,
		},
		"fullscreen": schema.BoolAttribute{
			Description: "Whether the custom page is in fullscreen mode.",
			Computed:    true,
		},
		"hidden": schema.BoolAttribute{
			Description: "Whether the custom page is hidden.",
			Computed:    true,
		},
		"revision": schema.Int64Attribute{
			Description: "The revision of the custom page.",
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: "The ID of the custom page.",
			Computed:    true,
		},
		"created_at": schema.StringAttribute{
			Description: "The date the custom page was created.",
			Computed:    true,
		},
		"updated_at": schema.StringAttribute{
			Description: "The date the custom page was last updated.",
			Computed:    true,
		},
		"algolia": schema.SingleNestedAttribute{
			Description: "Metadata about the Algolia search integration. " +
				"See <https://docs.readme.com/main/docs/search> for more information.",
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"publish_pending": schema.BoolAttribute{
					Computed: true,
				},
				"record_count": schema.Int64Attribute{
					Computed: true,
				},
				"updated_at": schema.StringAttribute{
					Computed: true,
				},
			},
		},
		"metadata": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"description": schema.StringAttribute{
					Computed: true,
				},
				"image": schema.ListAttribute{
					Computed:    true,
					ElementType: types.StringType,
				},
				"title": schema.StringAttribute{
					Computed: true,
				},
			},
		},
	}
}
