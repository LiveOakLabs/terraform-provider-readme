package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &categoryDataSource{}
	_ datasource.DataSourceWithConfigure = &categoryDataSource{}
)

// categoryDataSource is the data source implementation.
type categoryDataSource struct {
	client *readme.Client
}

// categoryModel maps an API specification to the apiSpecification schema data.
type categoryModel struct {
	CategoryType types.String `tfsdk:"category_type"`
	CreatedAt    types.String `tfsdk:"created_at"`
	ID           types.String `tfsdk:"id"`
	Order        types.Int64  `tfsdk:"order"`
	Project      types.String `tfsdk:"project"`
	Reference    types.Bool   `tfsdk:"reference"`
	Slug         types.String `tfsdk:"slug"`
	Title        types.String `tfsdk:"title"`
	Type         types.String `tfsdk:"type"`
	Version      types.String `tfsdk:"version"`
	VersionID    types.String `tfsdk:"version_id"`
}

// NewCategoryDataSource is a helper function to simplify the provider implementation.
func NewCategoryDataSource() datasource.DataSource {
	return &categoryDataSource{}
}

// Metadata returns the data source type name.
func (d *categoryDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_category"
}

// categoryCommonSchema returns the category schema that is common for the single category schema used in this data
// source and the list of categories used by the `readme_categories` data source.
var categoryCommonSchema = map[string]schema.Attribute{
	"created_at": schema.StringAttribute{
		Description: "Timestamp of when the version was created.",
		Computed:    true,
	},
	"id": schema.StringAttribute{
		Description: "The ID of the category.",
		Computed:    true,
	},
	"order": schema.Int64Attribute{
		Description: "The order of the category.",
		Computed:    true,
	},
	"project": schema.StringAttribute{
		Description: "The ID of the project the category is in.",
		Computed:    true,
	},
	"reference": schema.BoolAttribute{
		Description: "Indicates whether the category is a reference or not.",
		Computed:    true,
	},
	"slug": schema.StringAttribute{
		Description: "The slug of the category.",
		Required:    true,
	},
	"title": schema.StringAttribute{
		Description: "The title of the category.",
		Computed:    true,
	},
	"type": schema.StringAttribute{
		Description: "The category type.",
		Computed:    true,
	},
	"version_id": schema.StringAttribute{
		Description: "The version the category is associated with.",
		Computed:    true,
	},
}

// categorySingleSchema returns the Terraform data source schema for a single category, which adds a 'category_type' attribute
// to the common schema.
//
// This returns the common category schema with the `category_type` attribute, which is not used in the
// `readme_categories` data source.
func categorySingleSchema() map[string]schema.Attribute {
	mergedSchema := map[string]schema.Attribute{
		"category_type": schema.StringAttribute{
			Description: "The category type (different than 'type').",
			Computed:    true,
		},
		"version": schema.StringAttribute{
			Description: "The 'semver-ish' value of the version the category is under.",
			Computed:    true,
		},
	}

	// Merge the detail schema with the common schema.
	for k, v := range categoryCommonSchema {
		mergedSchema[k] = v
	}

	return mergedSchema
}

// Schema defines the schema for the data source.
func (d *categoryDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve metadata about a category on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getcategory> for more information about this API endpoint.\n\n",
		Attributes: categorySingleSchema(),
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *categoryDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state categoryModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get category metadata from ReadMe API
	category, apiResponse, err := d.client.Category.Get(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve category metadata.",
			clientError(err, apiResponse),
		)

		return
	}

	// Map response body to model
	state = categoryModel{
		CategoryType: types.StringValue(category.CategoryType),
		CreatedAt:    types.StringValue(category.CreatedAt),
		ID:           types.StringValue(category.ID),
		Order:        types.Int64Value(int64(category.Order)),
		Project:      types.StringValue(category.Project),
		Reference:    types.BoolValue(category.Reference),
		Slug:         types.StringValue(category.Slug),
		Title:        types.StringValue(category.Title),
		Type:         types.StringValue(category.Type),
		Version:      types.StringValue(versionClean(ctx, d.client, category.Version)),
		VersionID:    types.StringValue(category.Version),
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *categoryDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
