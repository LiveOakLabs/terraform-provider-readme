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
	_ datasource.DataSource              = &customPagesDataSource{}
	_ datasource.DataSourceWithConfigure = &customPagesDataSource{}
)

type customPagesDataSource struct {
	client *readme.Client
}

type customPagesDataSourceModel struct {
	ID      types.String                `tfsdk:"id"`
	Results []customPageDataSourceModel `tfsdk:"results"`
}

// NewCustomPagesDataSource is a helper function to simplify the provider implementation.
func NewCustomPagesDataSource() datasource.DataSource {
	return &customPagesDataSource{}
}

// Metadata returns the data source type name.
func (d *customPagesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_custom_pages"
}

// Read refreshes the Terraform state with the latest data.
func (d *customPagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customPagesDataSourceModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	results := []customPageDataSourceModel{}

	pages, _, err := d.client.CustomPage.GetAll()
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve custom pages.", err.Error())

		return
	}

	for _, page := range pages {
		results = append(results, customPageDatasourceMapToModel(page))
	}

	// state.Body = state.BodyClean
	state = customPagesDataSourceModel{
		ID:      types.StringValue("custom_pages"),
		Results: results,
	}

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *customPagesDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}

// Schema for the readme_custom_pages data source.
func (d *customPagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve custom pages from the ReadMe API.\n\n" +
			"See <https://docs.readme.com/reference/getcustompages> for more information about the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The state ID of the custom pages data source.",
				Computed:    true,
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: customPageDataSourceSchema(),
				},
			},
		},
	}
}
