// Package readme is a Terraform provider for interacting with the ReadMe.com API.
package readme

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces
var _ provider.Provider = &readmeProvider{}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &readmeProvider{
			Version: version,
		}
	}
}

// readmeProvider is the provider implementation.
type readmeProvider struct {
	Version string
}

// readmeProviderModel maps provider schema data to a Go type.
type readmeProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
	APIURL   types.String `tfsdk:"api_url"`
}

// Metadata returns the provider type name.
func (p *readmeProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "readme"
}

// Schema defines the provider-level schema for configuration data.
func (p *readmeProvider) Schema(
	ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "The ReadMe provider provides resources and data sources for interacting with the ReadMe.com API.",
		MarkdownDescription: "The ReadMe provider provides resources and data sources for interacting with the " +
			"[ReadMe.com](https://docs.readme.com/main/reference/intro-to-the-readme-api) API.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "Client token for accessing the ReadMe API. May alternatively be set with the " +
					"README_API_TOKEN environment variable.",
				MarkdownDescription: "Client token for accessing the ReadMe API. May alternatively be set with the " +
					"`README_API_TOKEN` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"api_url": schema.StringAttribute{
				Description: "URL for accessing the ReadMe API. May also be set with the README_API_URL " +
					"environment variable or left unset to use the default.",
				MarkdownDescription: "URL for accessing the ReadMe API. May also be set with the `README_API_URL` " +
					"environment variable or left unset to use the default.",
				Optional: true,
			},
		},
	}
}

// Configure prepares a readme API client for data sources and resources.
func (p *readmeProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	tflog.Info(ctx, fmt.Sprintf("Configuring ReadMe client version %s", p.Version))

	// Retrieve provider data from configuration
	var config readmeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	// Default values to environment variables, but override with Terraform configuration value if set.
	apiToken := os.Getenv("README_API_TOKEN")
	apiURL := os.Getenv("README_API_URL")

	// Use the config value if it's set
	if config.APIToken.ValueString() != "" {
		apiToken = config.APIToken.ValueString()
	}

	if config.APIURL.ValueString() != "" {
		apiURL = config.APIURL.ValueString()
	}

	// Ensure API token is set via the "api_token" attribute or README_API_TOKEN env var.
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing ReadMe API Token.",
			"The provider cannot create the Readme API client because there is a missing or empty value for the Readme API token. "+
				"Set the token value in the configuration or use the README_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// API URL is optional, but if it's specified, it must not be set to an empty value.
	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()

		if apiURL == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("api_url"),
				"Missing ReadMe API URL.",
				"The provider cannot create the Readme API client because the an API URL is set to an empty value. "+
					"Set the correct value in the configuration, use the README_API_URL environment variable, "+
					"or leave it unset to use the default value. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "api_token", apiToken)
	ctx = tflog.SetField(ctx, "api_url", apiURL)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "api_token")

	tflog.Debug(ctx, "Creating ReadMe client")

	var client *readme.Client
	var err error

	if apiURL == "" {
		client, err = readme.NewClient(apiToken)
	} else {
		client, err = readme.NewClient(apiToken, apiURL)
	}

	// Create a new Readme client using the configuration values
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Readme API Client.",
			"An unexpected error occurred when creating the Readme API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Readme Client Error: "+err.Error(),
		)

		return
	}

	// Make the Readme client available during DataSource and Resource type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured ReadMe client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *readmeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIRegistryDataSource,
		NewAPISpecificationDataSource,
		NewCategoriesDataSource,
		NewCategoryDataSource,
		NewCategoryDocsDataSource,
		NewDocDataSource,
		NewDocSearchDataSource,
		NewProjectDataSource,
		NewVersionDataSource,
		NewVersionsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *readmeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPISpecificationResource,
		NewCategoryResource,
		NewDocResource,
		NewImageResource,
		NewVersionResource,
	}
}

// versionModifier is a plan modifier that plans a change for an attribute if another specified attribute is changed.
type otherAttributeModifier struct {
	otherAttribute path.Path
}

// Description returns a plain text description of the modifier's behavior.
func (m otherAttributeModifier) Description(ctx context.Context) string {
	return "If another attribute is changed, this attribute will be changed."
}

// MarkdownDescription returns a markdown formatted description of the modifier's behavior.
func (m otherAttributeModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString implements a modifier for planning a change for an attribute if another specified attribute
// changes.
func (m otherAttributeModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	var otherPlanValue types.String
	var otherStateValue types.String

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, m.otherAttribute, &otherPlanValue)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, m.otherAttribute, &otherStateValue)...)

	// If the other attribute's value is unchanged, set the plan for this attribute to its current value (unchanged).
	if otherPlanValue == otherStateValue {
		resp.PlanValue = req.StateValue

		return
	}

	// If the other attribute is changed, mark this attribute as unknown.
	resp.PlanValue = types.StringUnknown()
}

// Custom plan modifier to flag an attribute for change if another specified attribute changes.
func changedIfOther(attribute path.Path) planmodifier.String {
	return otherAttributeModifier{
		otherAttribute: attribute,
	}
}

// boolPoint returns a pointer to a boolean.
func boolPoint(input bool) *bool {
	return &input
}

// intPoint returns a pointer to a boolean.
func intPoint(input int) *int {
	return &input
}

// apiRequestOptions returns options for making the API request with a version if a version is set.
// Otherwise, it returns an empty `readme.RequestOptions` struct.
func apiRequestOptions(version basetypes.StringValue) readme.RequestOptions {
	// If a version is provided, set the request option.
	var options readme.RequestOptions
	if version.ValueString() != "" {
		options = readme.RequestOptions{
			Version: version.ValueString(),
		}
	}

	return options
}

// versionClean returns the "clean" version for a version ID.
func versionClean(ctx context.Context, client *readme.Client, versionID string) string {
	version, apiResponse, err := client.Version.Get("id:" + versionID)
	if err != nil {
		tflog.Info(
			ctx,
			fmt.Sprintf("error resolving version: %s. API response: %+v", err.Error(), apiResponse),
		)

		return ""
	}

	if version.VersionClean == "" {
		tflog.Info(ctx, "the version returned is empty")

		return ""
	}

	return version.VersionClean
}

// clientError is a helper function for formatting a Terraform diagnostics error response string
// from the client library and API.
//
// It accepts the raw error and APIResponse struct. If the APIResponse includes an error message,
// it will be appended to the error with a line break.
// Functions that make an API request should use the returned string as the second argument to the
// Terraform diagnostics AddError() function, which is used as the detailed message in a Terraform
// error.
func clientError(err error, apiResponse *readme.APIResponse) string {
	diagErr := err.Error()

	if apiResponse != nil && apiResponse.APIErrorResponse.Message != "" {
		diagErr += fmt.Sprintf("\nAPI Error Message: " + apiResponse.APIErrorResponse.Message)
	}

	return diagErr
}
