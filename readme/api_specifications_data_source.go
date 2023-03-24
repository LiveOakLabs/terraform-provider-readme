package readme

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &apiSpecificationsDataSource{}
	_ datasource.DataSourceWithConfigure = &apiSpecificationsDataSource{}
)

// apiSpecificationsDataSource is the data source implementation.
type apiSpecificationsDataSource struct {
	client *readme.Client
}

// apiSpecificationDataSourceModel maps an API specification to the apiSpecification schema data.
type apiSpecificationsDataSourceModel struct {
	Filter *apiSpecificationsDataSourceFilter `tfsdk:"filter"`
	ID     types.String                       `tfsdk:"id"`
	Specs  []apiSpecificationsItemModel       `tfsdk:"specs"`
	SortBy types.String                       `tfsdk:"sort_by"`
}

type apiSpecificationsItemModel struct {
	Category   types.Object `tfsdk:"category"`
	ID         types.String `tfsdk:"id"`
	LastSynced types.String `tfsdk:"last_synced"`
	Source     types.String `tfsdk:"source"`
	Title      types.String `tfsdk:"title"`
	Type       types.String `tfsdk:"type"`
	Version    types.String `tfsdk:"version"`
}

// apiSpecificationsDataSourceFilter is the filter schema for the apiSpecification data source.
type apiSpecificationsDataSourceFilter struct {
	CategoryID    []types.String `tfsdk:"category_id"`
	CategorySlug  []types.String `tfsdk:"category_slug"`
	CategoryTitle []types.String `tfsdk:"category_title"`
	HasCategory   types.Bool     `tfsdk:"has_category"`
	Title         []types.String `tfsdk:"title"`
	Version       []types.String `tfsdk:"version"`
}

// NewAPISpecificationDataSource is a helper function to simplify the provider implementation.
func NewAPISpecificationsDataSource() datasource.DataSource {
	return &apiSpecificationsDataSource{}
}

// Metadata returns the data source type name.
func (d *apiSpecificationsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_api_specifications"
}

// Configure adds the provider configured client to the data source.
func (d *apiSpecificationsDataSource) Configure(
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
func (d *apiSpecificationsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve multiple API specifications from ReadMe. " +
			"The `filter` attribute may be used to filter API specifications by category ID, category slug, category " +
			"title, title, version, or whether or not the API specifications have a category. " +
			"See <https://docs.readme.com/main/reference/getapispecification> for more information about this API " +
			"endpoint.",
		Attributes: map[string]schema.Attribute{
			"filter": schema.SingleNestedAttribute{
				Description: "Filter API specifications by the specified criteria. Ommitting this attribute will " +
					"return all API specifications. All category filters are 'OR' filters except `has_category`, which " +
					"works as an 'AND' filter with the other category filters.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"category_id": schema.ListAttribute{
						Description: "Return API specifications matching the specified category IDs.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"category_slug": schema.ListAttribute{
						Description: "Return API specifications matching the specified category slugs.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"category_title": schema.ListAttribute{
						Description: "Return API specifications matching the specified category titles.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"has_category": schema.BoolAttribute{
						Description: "Return only API specifications with a category. This attribute works as an 'AND' " +
							"filter with the other category filters.",
						Optional: true,
					},
					"title": schema.ListAttribute{
						Description: "Return API specifications matching the specified title.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"version": schema.ListAttribute{
						Description: "Return API specifications matching the specified version ID.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"id": schema.StringAttribute{
				Description: "ID for the data source.",
				Computed:    true,
			},
			"specs": schema.ListNestedAttribute{
				Description: "List of API specifications",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
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
				},
			},
			"sort_by": schema.StringAttribute{
				Description: "Sort the returned API specifications by the specified key." +
					"Valid values are `title` or `last_synced`. If unset, API specifications are in the order they were " +
					"returned by the API.",
				Optional: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *apiSpecificationsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var (
		state       apiSpecificationsDataSourceModel
		apiResponse *readme.APIResponse
		err         error
	)

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all API specifications.
	apiSpecs, apiResponse, err := d.client.APISpecification.GetAll()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve API specifications.",
			clientError(err, apiResponse),
		)

		return
	}

	// Optionally sort API specifications.
	if !state.SortBy.IsNull() {
		apiSpecs, err = sortAPISpecifications(apiSpecs, state.SortBy.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve API specifications.",
				fmt.Sprintf("Unable to sort API specifications. %s", err.Error()),
			)

			return
		}
	}

	// Find a matching API specification.
	for _, spec := range apiSpecs {
		if !specMatchesMultiFilters(ctx, state.Filter, spec) {
			continue
		}

		// Map response body to model.
		specModel := apiSpecificationsItemModel{
			ID:         types.StringValue(spec.ID),
			LastSynced: types.StringValue(spec.LastSynced),
			Source:     types.StringValue(spec.Source),
			Title:      types.StringValue(spec.Title),
			Type:       types.StringValue(spec.Type),
			Version:    types.StringValue(spec.Version),
			Category:   specCategoryObject(spec),
		}

		state.Specs = append(state.Specs, specModel)
	}

	// Map response body to model.
	state = apiSpecificationsDataSourceModel{
		Filter: state.Filter,
		ID:     types.StringValue("readme_api_specifications"),
		Specs:  state.Specs,
		SortBy: state.SortBy,
	}

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// sortAPISpecifications sorts the API specifications by the specified key.
func sortAPISpecifications(specs []readme.APISpecification, sortBy string) ([]readme.APISpecification, error) {
	switch sortBy {
	case "title":
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].Title < specs[j].Title
		})
	case "last_synced":
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].LastSynced < specs[j].LastSynced
		})
	default:
		return nil, fmt.Errorf("invalid sort value: %s", sortBy)
	}

	return specs, nil
}

// specMatchesMultiFilters returns true if the API specification matches the filter criteria.
// This uses the plural form of the filter attributes to allow for multiple values to be specified.
func specMatchesMultiFilters(
	ctx context.Context,
	filter *apiSpecificationsDataSourceFilter,
	spec readme.APISpecification,
) bool {
	// If no filter is specified, return true.
	if filter == nil {
		return true
	}

	// Filter by category visibility.
	passCategory := !filter.HasCategory.ValueBool() || (filter.HasCategory.ValueBool() && spec.Category.ID != "")

	if !passCategory {
		tflog.Info(ctx, fmt.Sprintf("Skipping API specification %s due to category visibility.", spec.ID))
		return false
	}

	// If has_category but none of the other filters are specified, return true.
	if filter.Title == nil &&
		filter.Version == nil &&
		filter.CategoryID == nil &&
		filter.CategorySlug == nil &&
		filter.CategoryTitle == nil {
		return true
	}

	// Filter by spec title.
	for _, title := range filter.Title {
		if spec.Title == title.ValueString() {
			tflog.Info(ctx, fmt.Sprintf("API specification %s matched title filter.", spec.Title))
			return true
		}
	}

	// Filter by spec version.
	for _, version := range filter.Version {
		if spec.Version == version.ValueString() {
			tflog.Info(ctx, fmt.Sprintf("API specification %s matched version filter.", spec.Version))
			return true
		}
	}

	// Filter by category ID.
	for _, categoryID := range filter.CategoryID {
		if passCategory && spec.Category.ID == categoryID.ValueString() {
			tflog.Info(ctx, fmt.Sprintf("API specification %s matched category ID filter.", spec.Category.ID))
			return true
		}
	}

	// Filter by category slug.
	for _, categorySlug := range filter.CategorySlug {
		if passCategory && spec.Category.Slug == categorySlug.ValueString() {
			tflog.Info(ctx, fmt.Sprintf("API specification %s matched category slug filter.", spec.Category.Slug))
			return true
		}
	}

	// Filter by category title.
	for _, categoryTitle := range filter.CategoryTitle {
		if passCategory && spec.Category.Title == categoryTitle.ValueString() {
			tflog.Info(
				ctx,
				fmt.Sprintf(
					"API specification %s category %s matched category title filter.", spec.Title, spec.Category.Title,
				),
			)
			return true
		}
	}

	tflog.Info(ctx, "No filter criteria matched, returning false.")

	return false
}
