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
	_ datasource.DataSource              = &categoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &categoriesDataSource{}
)

// categoriesDataSource is the data source implementation.
type categoriesDataSource struct {
	client *readme.Client
}

type categoriesModel struct {
	ID         types.String        `tfsdk:"id"`
	Categories []categoryListModel `tfsdk:"categories"`
}

type categoryListModel struct {
	CreatedAt types.String `tfsdk:"created_at"`
	ID        types.String `tfsdk:"id"`
	Order     types.Int64  `tfsdk:"order"`
	Project   types.String `tfsdk:"project"`
	Reference types.Bool   `tfsdk:"reference"`
	Slug      types.String `tfsdk:"slug"`
	Title     types.String `tfsdk:"title"`
	Type      types.String `tfsdk:"type"`
	VersionID types.String `tfsdk:"version_id"`
}

// NewCategoriesDataSource is a helper function to simplify the provider implementation.
func NewCategoriesDataSource() datasource.DataSource {
	return &categoriesDataSource{}
}

// Metadata returns the data source type name.
func (d *categoriesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_categories"
}

// Schema defines the schema for the data source.
func (d *categoriesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve all categories for a project on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getcategories> for more information about this API endpoint.\n\n",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The internal Terraform ID of the categories data source.",
				Computed:    true,
			},
			"categories": schema.ListNestedAttribute{
				Description: "List of category summaries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: categoryCommonSchema,
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *categoriesDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state categoriesModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get categories metadata from ReadMe API.
	categories, apiResponse, err := d.client.Category.GetAll()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve categories metadata.",
			clientError(err, apiResponse),
		)

		return
	}

	for _, category := range categories {
		// Map response body to model.
		cat := categoryListModel{
			CreatedAt: types.StringValue(category.CreatedAt),
			ID:        types.StringValue(category.ID),
			Order:     types.Int64Value(int64(category.Order)),
			Project:   types.StringValue(category.Project),
			Reference: types.BoolValue(category.Reference),
			Slug:      types.StringValue(category.Slug),
			Title:     types.StringValue(category.Title),
			Type:      types.StringValue(category.Type),
			VersionID: types.StringValue(category.Version),
		}

		state.Categories = append(state.Categories, cat)
	}

	// The ID attribute is only used by Terraform and the provider internally.
	// It doesn't exist in the API for the list of categoriess but is required to be set.
	state.ID = types.StringValue("readme_categories")

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *categoriesDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
