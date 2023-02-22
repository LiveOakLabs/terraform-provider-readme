package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &apiRegistryDataSource{}
	_ datasource.DataSourceWithConfigure = &apiRegistryDataSource{}
)

// registryDataSource is the data source implementation.
type apiRegistryDataSource struct {
	client *readme.Client
}

// registryModel maps an API specification to the apiSpecification schema data.
type apiRegistryModel struct {
	Definition types.String `tfsdk:"definition"`
	ID         types.String `tfsdk:"id"`
	UUID       types.String `tfsdk:"uuid"`
}

// NewAPIRegistryDataSource is a helper function to simplify the provider implementation.
func NewAPIRegistryDataSource() datasource.DataSource {
	return &apiRegistryDataSource{}
}

// Metadata returns the data source type name.
func (d *apiRegistryDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_api_registry"
}

// Schema defines the schema for the data source.
func (d *apiRegistryDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve an API specification definition from the API registry on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getapiregistry> for more information about this API endpoint.",
		Attributes: map[string]schema.Attribute{
			"definition": schema.StringAttribute{
				Description: "The raw JSON definition of an API specification.",
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "The UUID of an API registry definition.",
				Required:    true,
			},
			// The 'id' isn't returned by ReadMe - it's for Terraform use.
			// See https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
			"id": schema.StringAttribute{
				Description: "The internal ID of this resource.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *apiRegistryDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state apiRegistryModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	uuid := state.UUID.ValueString()
	if uuid == "" {
		resp.Diagnostics.AddError("Unable to retrieve API registry.",
			"You must provide the 'uuid' attribute to query the API registry.\n"+
				"See https://docs.readme.com/main/reference/getapiregistry for more information.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the specification from the API registry
	registry, apiResponse, err := d.client.APIRegistry.Get(uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve API registry metadata.",
			clientError(err, apiResponse),
		)

		return
	}

	// Map response body to model
	state = apiRegistryModel{
		Definition: types.StringValue(registry),
		UUID:       types.StringValue(uuid),
	}

	state.ID = types.StringValue("readme")

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *apiRegistryDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
