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
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

// projectDataSource is the data source implementation.
type projectDataSource struct {
	client *readme.Client
}

// projectModel maps an API specification to the apiSpecification schema data.
type projectModel struct {
	Name      types.String `tfsdk:"name"`
	SubDomain types.String `tfsdk:"subdomain"`
	JWTSecret types.String `tfsdk:"jwt_secret"`
	BaseURL   types.String `tfsdk:"base_url"`
	Plan      types.String `tfsdk:"plan"`
	ID        types.String `tfsdk:"id"`
}

// NewProjectDataSource is a helper function to simplify the provider implementation.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// Metadata returns the data source type name.
func (d *projectDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source.
func (d *projectDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve metadata about a project on ReadMe.com\n\n" +
			"The project metadata associated with the API token is returned.\n\n" +
			"See <https://docs.readme.com/main/reference/getproject> for more information about this API endpoint.\n\n",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the project.",
				Computed:    true,
			},
			"subdomain": schema.StringAttribute{
				Description: "The base URL for the project.",
				Computed:    true,
			},
			"jwt_secret": schema.StringAttribute{
				Description: "JWT Secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "The base URL of the project.",
				Computed:    true,
			},
			"plan": schema.StringAttribute{
				Description: "The account subscription plan.",
				Computed:    true,
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
func (d *projectDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state projectModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project metadata from ReadMe API
	project, apiResponse, err := d.client.Project.Get()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve project metadata.",
			clientError(err, apiResponse),
		)

		return
	}

	// Map response body to model
	state = projectModel{
		Name:      types.StringValue(project.Name),
		SubDomain: types.StringValue(project.SubDomain),
		JWTSecret: types.StringValue(project.JWTSecret),
		BaseURL:   types.StringValue(project.BaseURL),
		Plan:      types.StringValue(project.Plan),
	}

	// The ID isn't returned in the data source but is tracked internally and
	// required for testing with the Terraform SDK.
	// See https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	state.ID = types.StringValue("readme")

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *projectDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
