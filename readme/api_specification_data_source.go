package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &apiSpecificationDataSource{}
	_ datasource.DataSourceWithConfigure = &apiSpecificationDataSource{}
)

// apiSpecificationDataSource is the data source implementation.
type apiSpecificationDataSource struct {
	client *readme.Client
}

// apiSpecificationDataSourceModel maps an API specification to the apiSpecification schema data.
type apiSpecificationDataSourceModel struct {
	Category   types.Object                      `tfsdk:"category"`
	ID         types.String                      `tfsdk:"id"`
	LastSynced types.String                      `tfsdk:"last_synced"`
	Source     types.String                      `tfsdk:"source"`
	Title      types.String                      `tfsdk:"title"`
	Type       types.String                      `tfsdk:"type"`
	Version    types.String                      `tfsdk:"version"`
	Filter     *apiSpecificationDataSourceFilter `tfsdk:"filter"`
}

// apiSpecificationDataSourceFilter is the filter schema for the apiSpecification data source.
type apiSpecificationDataSourceFilter struct {
	CategoryID    types.String `tfsdk:"category_id"`
	CategorySlug  types.String `tfsdk:"category_slug"`
	CategoryTitle types.String `tfsdk:"category_title"`
	HasCategory   types.Bool   `tfsdk:"has_category"`
}

// NewAPISpecificationDataSource is a helper function to simplify the provider implementation.
func NewAPISpecificationDataSource() datasource.DataSource {
	return &apiSpecificationDataSource{}
}

// Metadata returns the data source type name.
func (d *apiSpecificationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_api_specification"
}

// Configure adds the provider configured client to the data source.
func (d *apiSpecificationDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}

// Schema defines the schema for the data source Terraform attributes.
func (d *apiSpecificationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	// nolint: goconst // Attribute descriptions are reused.
	resp.Schema = schema.Schema{
		Description: "Retrieve metadata about an API specification on ReadMe.com\n\n" +
			"An ID or title must be specified to retrieve an API specification. The `filter` attribute may be used " +
			"to filter API specifications by category ID, category slug, category title, or whether or not the API " +
			"specification has a category. " +
			"See <https://docs.readme.com/main/reference/getapispecification> for more information about this API " +
			"endpoint.",
		Attributes: map[string]schema.Attribute{
			"category": schema.ObjectAttribute{
				Description: "Category information",
				Computed:    true,
				AttributeTypes: map[string]attr.Type{
					"id":    types.StringType,
					"slug":  types.StringType,
					"order": types.Int64Type,
					"title": types.StringType,
					"type":  types.StringType,
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the API specification.",
				Computed:    true,
				Optional:    true,
			},
			"filter": schema.SingleNestedAttribute{
				Description: "Filter API specifications by the specified criteria.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"category_id": schema.StringAttribute{
						Description: "Return only API specifications with the specified category ID.",
						Optional:    true,
					},
					"category_slug": schema.StringAttribute{
						Description: "Return only API specifications with the specified category slug.",
						Optional:    true,
					},
					"category_title": schema.StringAttribute{
						Description: "Return only API specifications with the specified category title.",
						Optional:    true,
					},
					"has_category": schema.BoolAttribute{
						Description: "Return only API specifications with a category.",
						Optional:    true,
					},
				},
			},
			"last_synced": schema.StringAttribute{
				Description: "Timestamp of last synchronization.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The creation source of the API specification.",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "The title of the API specification derived from the specification JSON.",
				Computed:    true,
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the API specification.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The version of the API specification.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *apiSpecificationDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state apiSpecificationDataSourceModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no ID or title is specified, return an error.
	if state.ID.ValueString() == "" && state.Title.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Unable to retrieve API specification metadata.",
			"An ID or title must be specified to retrieve an API specification.",
		)

		return
	}

	var (
		apiSpec     readme.APISpecification
		apiResponse *readme.APIResponse
		err         error
	)

	// If an ID is specified, use that to retrieve the API specification.
	if state.ID.ValueString() != "" { // nolint: nestif // TODO: Refactor.
		// Get API specification by ID.
		apiSpec, apiResponse, err = d.client.APISpecification.Get(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve API specification metadata.",
				clientError(err, apiResponse),
			)

			return
		}

		// Filter the specification by criteria.
		if !specMatchesFilter(state, apiSpec) {
			resp.Diagnostics.AddError(
				"Unable to find API specification.",
				"Unable to find API specification with the specified criteria.",
			)

			return
		}
	} else {
		// Get all API specifications.
		apiSpecs, apiResponse, err := d.client.APISpecification.GetAll()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve API specifications.",
				clientError(err, apiResponse),
			)

			return
		}

		if len(apiSpecs) == 0 {
			resp.Diagnostics.AddError(
				"No API specifications found.",
				"No API specifications were found in the ReadMe project when searching by title.",
			)

			return
		}

		// Find a matching API specification.
		for _, spec := range apiSpecs {
			if !specMatchesFilter(state, spec) {
				continue
			}

			apiSpec = spec

			break
		}

		// If no matching API specification was found, return an error.
		if apiSpec.ID == "" {
			resp.Diagnostics.AddError(
				"Unable to find API specification with title: "+state.Title.ValueString(), "",
			)

			return
		}
	}

	// Map response body to model.
	state = apiSpecificationDataSourceModel{
		ID:         types.StringValue(apiSpec.ID),
		LastSynced: types.StringValue(apiSpec.LastSynced),
		Source:     types.StringValue(apiSpec.Source),
		Title:      types.StringValue(apiSpec.Title),
		Type:       types.StringValue(apiSpec.Type),
		Version:    types.StringValue(apiSpec.Version),
		Category:   specCategoryObject(apiSpec),
		Filter:     state.Filter,
	}

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// specMatchesFilter returns true if the API specification matches the filter criteria.
func specMatchesFilter(state apiSpecificationDataSourceModel, spec readme.APISpecification) bool {
	// Title is specified and does not match.
	if !state.Title.IsNull() && spec.Title != state.Title.ValueString() {
		return false
	}

	// If no additional filter is specified, return true.
	if state.Filter == nil {
		return true
	}

	// Filter by category ID.
	if !state.Filter.CategoryID.IsNull() && spec.Category.ID != state.Filter.CategoryID.ValueString() {
		return false
	}

	// Filter by category slug.
	if !state.Filter.CategorySlug.IsNull() && spec.Category.Slug != state.Filter.CategorySlug.ValueString() {
		return false
	}

	// Filter by category title.
	if !state.Filter.CategoryTitle.IsNull() && spec.Category.Title != state.Filter.CategoryTitle.ValueString() {
		return false
	}

	// Filter by category visibility.
	if state.Filter.HasCategory.ValueBool() && spec.Category.ID == "" {
		return false
	}

	return true
}
