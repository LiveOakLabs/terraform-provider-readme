package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &versionDataSource{}
	_ datasource.DataSourceWithConfigure = &versionDataSource{}
)

// versionDataSource is the data source implementation.
type versionDataSource struct {
	client *readme.Client
}

// versionDetail maps a ReadMe Version to the Terraform schema.
type versionDetail struct {
	Categories   types.List   `tfsdk:"categories"`
	Codename     types.String `tfsdk:"codename"`
	CreatedAt    types.String `tfsdk:"created_at"`
	ForkedFrom   types.String `tfsdk:"forked_from"`
	ID           types.String `tfsdk:"id"`
	IsBeta       types.Bool   `tfsdk:"is_beta"`
	IsDeprecated types.Bool   `tfsdk:"is_deprecated"`
	IsHidden     types.Bool   `tfsdk:"is_hidden"`
	IsStable     types.Bool   `tfsdk:"is_stable"`
	Project      types.String `tfsdk:"project"`
	ReleaseDate  types.String `tfsdk:"release_date"`
	Version      types.String `tfsdk:"version"`
	VersionClean types.String `tfsdk:"version_clean"`
}

// NewVersionDataSource is a helper function to simplify the provider implementation.
func NewVersionDataSource() datasource.DataSource {
	return &versionDataSource{}
}

// Metadata returns the data source type name.
func (d *versionDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

// Schema defines the schema for the data source.
func (d *versionDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve version metadata from the ReadMe.com API.\n\n" +
			"See <https://docs.readme.com/main/reference/getversion> for more information about this API endpoint.",
		Attributes: versionDetailSchema(),
	}
}

// versionDetailSchema returns the schema attributes for a specific versions's detail.
// This merges the schema from versionSchema() to return the full schema.
func versionDetailSchema() map[string]schema.Attribute {
	commonSchema := versionSchema()
	details := map[string]schema.Attribute{
		"categories": schema.ListAttribute{
			Description: "List of category IDs the version is associated with.",
			Computed:    true,
			ElementType: types.StringType,
		},
		"forked_from": schema.StringAttribute{
			Description: "ID of the version that was forked from.",
			Computed:    true,
		},
		"project": schema.StringAttribute{
			Description: "The project the version is in.",
			Computed:    true,
		},
		"release_date": schema.StringAttribute{
			Description: "Timestamp of when the version was released.",
			Computed:    true,
		},
	}

	// Merge the detail schema with the common schema.
	for k, v := range details {
		commonSchema[k] = v
	}

	return commonSchema
}

// versionSchema returns the schema attributes that are common between a version's details and the summary returned
// in the list of all versions.
func versionSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"codename": schema.StringAttribute{
			Description: "Dubbed name of version.",
			Computed:    true,
		},
		"created_at": schema.StringAttribute{
			Description: "Timestamp of when the version was created.",
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: "The ID of the version.",
			Computed:    true,
			Optional:    true,
		},
		"is_beta": schema.BoolAttribute{
			Description: "Indicates if the version is beta.",
			Computed:    true,
		},
		"is_deprecated": schema.BoolAttribute{
			Description: "Indicates if the version is deprecated.",
			Computed:    true,
		},
		"is_hidden": schema.BoolAttribute{
			Description: "Indicates if the version is hidden.",
			Computed:    true,
		},
		"is_stable": schema.BoolAttribute{
			Description: "Indicates if the version is stable.",
			Computed:    true,
		},
		"version": schema.StringAttribute{
			Description: "The version string, usually a semantic version.",
			Computed:    true,
			Optional:    true,
		},
		"version_clean": schema.StringAttribute{
			Description: "A 'clean' version string with certain characters replaced, usually a semantic version.",
			Computed:    true,
			Optional:    true,
		},
	}
}

// Read the remote state and refresh the Terraform state with the latest data.
func (d *versionDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state versionDetail

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if state.Version.IsNull() && state.VersionClean.IsNull() && state.ID.IsNull() {
		resp.Diagnostics.AddError("'id', 'version', or 'version_clean' must be set.", "")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the version parameter to query.
	// In order of preference: version_clean, version, id.
	var reqVersion string
	if state.VersionClean.ValueString() != "" {
		reqVersion = state.VersionClean.ValueString()
	} else if state.Version.ValueString() != "" {
		reqVersion = state.Version.ValueString()
	} else if state.ID.ValueString() != "" {
		reqVersion = "id:" + state.ID.ValueString()
	}

	// Get the version.
	version, apiResponse, err := d.client.Version.Get(reqVersion)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read version.", clientError(err, apiResponse))

		return
	}

	// Map response body to model.
	state = versionDetail{
		Codename:     types.StringValue(version.Codename),
		CreatedAt:    types.StringValue(version.CreatedAt),
		ForkedFrom:   types.StringValue(version.ForkedFrom),
		ID:           types.StringValue(version.ID),
		IsBeta:       types.BoolValue(version.IsBeta),
		IsDeprecated: types.BoolValue(version.IsDeprecated),
		IsHidden:     types.BoolValue(version.IsHidden),
		IsStable:     types.BoolValue(version.IsStable),
		Project:      types.StringValue(version.Project),
		ReleaseDate:  types.StringValue(version.ReleaseDate),
		Version:      types.StringValue(version.Version),
		VersionClean: types.StringValue(version.VersionClean),
	}

	// Map category list to Terraform type.
	categories := make([]attr.Value, 0, len(version.Categories))
	for _, cat := range version.Categories {
		categories = append(categories, types.StringValue(cat))
	}
	state.Categories, diags = types.ListValue(types.StringType, categories)
	resp.Diagnostics.Append(diags...)

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *versionDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
