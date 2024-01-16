package readme

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &changelogDataSource{}
	_ datasource.DataSourceWithConfigure = &changelogDataSource{}
)

type changelogDataSource struct {
	client *readme.Client
}

// changelogDataSourceModel is the data source model used by the
// readme_changelog and readme_changelogs data sources.
type changelogDataSourceModel struct {
	Algolia   types.Object `tfsdk:"algolia"`
	Body      types.String `tfsdk:"body"`
	CreatedAt types.String `tfsdk:"created_at"`
	HTML      types.String `tfsdk:"html"`
	Hidden    types.Bool   `tfsdk:"hidden"`
	ID        types.String `tfsdk:"id"`
	Metadata  *docMetadata `tfsdk:"metadata"`
	Revision  types.Int64  `tfsdk:"revision"`
	Slug      types.String `tfsdk:"slug"`
	Title     types.String `tfsdk:"title"`
	Type      types.String `tfsdk:"type"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// changelogDatasourceMapToModel maps a readme.CustomPage to a changelogDataSourceModel
// for use in the readme_changelog and readme_changelogs data sources.
func changelogDatasourceMapToModel(page readme.Changelog) changelogDataSourceModel {
	image := []string{}
	if page.Metadata.Image != nil {
		for _, i := range page.Metadata.Image {
			image = append(image, fmt.Sprintf("%v", i))
		}
	}

	model := changelogDataSourceModel{
		Title:     types.StringValue(page.Title),
		Type:      types.StringValue(page.Type),
		Slug:      types.StringValue(page.Slug),
		Body:      types.StringValue(page.Body),
		HTML:      types.StringValue(page.HTML),
		Hidden:    types.BoolValue(page.Hidden),
		Revision:  types.Int64Value(int64(page.Revision)),
		ID:        types.StringValue(page.ID),
		CreatedAt: types.StringValue(page.CreatedAt),
		UpdatedAt: types.StringValue(page.UpdatedAt),
		Metadata: &docMetadata{
			Image:       image,
			Title:       page.Metadata.Title,
			Description: page.Metadata.Description,
		},
		Algolia: docModelAlgoliaValue(page.Algolia),
	}

	return model
}

// NewChangelogDataSource is a helper function to simplify the provider implementation.
func NewChangelogDataSource() datasource.DataSource {
	return &changelogDataSource{}
}

// Metadata returns the data source type name.
func (d *changelogDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_changelog"
}

// Read refreshes the Terraform state with the latest data.
func (d *changelogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state changelogDataSourceModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, _, err := d.client.Changelog.Get(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve changelog.", err.Error())

		return
	}

	state = changelogDatasourceMapToModel(page)

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *changelogDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}

// Schema for the readme_changelog data source.
func (d *changelogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve a changelog from the ReadMe API.\n\n" +
			"See <https://docs.readme.com/reference/getchangelog> for more information about the API.",
		// nolint:goconst
		Attributes: map[string]schema.Attribute{
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
			"body": schema.StringAttribute{
				Description: "The body of the changelog.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The date the changelog was created.",
				Computed:    true,
			},
			"hidden": schema.BoolAttribute{
				Description: "Whether the changelog is hidden.",
				Computed:    true,
			},
			"html": schema.StringAttribute{
				Description: "The HTML of the changelog.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the changelog.",
				Computed:    true,
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
			"revision": schema.Int64Attribute{
				Description: "The revision of the changelog.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the changelog.",
				Required:    true,
			},
			"title": schema.StringAttribute{
				Description: "The title of the changelog.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the changelog.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The date the changelog was last updated.",
				Computed:    true,
			},
		},
	}
}
