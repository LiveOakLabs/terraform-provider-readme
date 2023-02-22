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
	_ datasource.DataSource              = &versionsDataSource{}
	_ datasource.DataSourceWithConfigure = &versionsDataSource{}
)

// versionsDataSource is the data source implementation.
type versionsDataSource struct {
	client *readme.Client
}

// versionsList maps a ReadMe Version to the Terraform schema.
type versionsList struct {
	ID       types.String     `tfsdk:"id"`
	Versions []versionSummary `tfsdk:"versions"`
}

// versionSummary maps a version in the list of all versions to the Terraform schema.
type versionSummary struct {
	Codename     types.String `tfsdk:"codename"`
	CreatedAt    types.String `tfsdk:"created_at"`
	ForkedFrom   types.String `tfsdk:"forked_from"`
	ID           types.String `tfsdk:"id"`
	IsBeta       types.Bool   `tfsdk:"is_beta"`
	IsDeprecated types.Bool   `tfsdk:"is_deprecated"`
	IsHidden     types.Bool   `tfsdk:"is_hidden"`
	IsStable     types.Bool   `tfsdk:"is_stable"`
	Version      types.String `tfsdk:"version"`
	VersionClean types.String `tfsdk:"version_clean"`
}

// NewVersionsDataSource is a helper function to simplify the provider implementation.
func NewVersionsDataSource() datasource.DataSource {
	return &versionsDataSource{}
}

// Metadata returns the data source type name.
func (d *versionsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_versions"
}

// versionSummarySchema returns the schema attributes for a version list summary.
// This merges the schema from versionSchema() to return the full schema.
// Refer to 'version_data_source' for the common versionSchema.
func versionSummarySchema() map[string]schema.Attribute {
	commonSchema := versionSchema()
	summary := map[string]schema.Attribute{
		"forked_from": schema.StringAttribute{
			Description: "ID of the version that was forked from.",
			Computed:    true,
		},
	}
	for x, y := range summary {
		commonSchema[x] = y
	}

	return commonSchema
}

// Schema defines the schema for the data source.
func (d *versionsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve a list of versions with metadata from ReadMe.\n\n" +
			"See <https://docs.readme.com/main/reference/getversions> for more information about this API endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internally used identifier attribute. This attribute only exists within the provider " +
					"not the API.",
				Computed: true,
			},
			"versions": schema.ListNestedAttribute{
				Description: "The list of versions on ReadMe.com.",
				Computed:    true,
				Optional:    false,
				NestedObject: schema.NestedAttributeObject{
					Attributes: versionSummarySchema(),
				},
			},
		},
	}
}

// Read the remote state and refresh the Terraform state with the latest data.
func (d *versionsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state versionsList

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Versions list.
	versions, apiResponse, err := d.client.Version.GetAll()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read versions.", clientError(err, apiResponse))

		return
	}

	for _, vers := range versions {
		// Map response body to model.
		version := versionSummary{
			Codename:     types.StringValue(vers.Codename),
			CreatedAt:    types.StringValue(vers.CreatedAt),
			ForkedFrom:   types.StringValue(vers.ForkedFrom),
			ID:           types.StringValue(vers.ID),
			IsBeta:       types.BoolValue(vers.IsBeta),
			IsDeprecated: types.BoolValue(vers.IsDeprecated),
			IsHidden:     types.BoolValue(vers.IsHidden),
			IsStable:     types.BoolValue(vers.IsStable),
			Version:      types.StringValue(vers.Version),
			VersionClean: types.StringValue(vers.VersionClean),
		}

		state.Versions = append(state.Versions, version)
	}

	// The ID attribute is only used by Terraform and the provider internally.
	// It doesn't exist in the API for the list of versions but is required to
	// be set.
	state.ID = types.StringValue("readme_versions")

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *versionsDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
