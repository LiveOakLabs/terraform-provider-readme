package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &customPageDataSource{}
	_ datasource.DataSourceWithConfigure = &customPageDataSource{}
)

type customPageDataSource struct {
	client *readme.Client
}

// NewCustomPageDataSource is a helper function to simplify the provider implementation.
func NewCustomPageDataSource() datasource.DataSource {
	return &customPageDataSource{}
}

// Metadata returns the data source type name.
func (d *customPageDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_custom_page"
}

// Read refreshes the Terraform state with the latest data.
func (d *customPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customPageDataSourceModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, _, err := d.client.CustomPage.Get(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve custom pages.", err.Error())

		return
	}

	state = customPageDatasourceMapToModel(page)

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *customPageDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}

// Schema for the readme_custom_page data source.
func (d *customPageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := customPageDataSourceSchema()
	attrs["slug"] = schema.StringAttribute{
		Description: "The slug of the custom page.",
		Required:    true,
	}
	resp.Schema = schema.Schema{
		Description: "Retrieve a custom page from the ReadMe API.\n\n" +
			"See <https://docs.readme.com/reference/getcustompages> for more information about the API.",
		Attributes: attrs,
	}
}
