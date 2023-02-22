package readme

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &categoryResource{}
	_ resource.ResourceWithConfigure   = &categoryResource{}
	_ resource.ResourceWithImportState = &categoryResource{}
)

// categoryResource is the data source implementation.
type categoryResource struct {
	client *readme.Client
}

// NewCategoryResource is a helper function to simplify the provider
// implementation.
func NewCategoryResource() resource.Resource {
	return &categoryResource{}
}

// Metadata returns the data source type name.
func (r *categoryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_category"
}

// Configure adds the provider configured client to the data source.
func (r *categoryResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*readme.Client)
}

// ValidateConfig is used for validating attribute values.
func (r categoryResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data categoryModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.Type.ValueString() != "guide" && data.Type.ValueString() != "reference" {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Category type must be 'guide' or 'reference'.", "",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *categoryResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manage categories for a project on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getcategory> for more information about this API endpoint.\n\n",
		Attributes: map[string]schema.Attribute{
			"category_type": schema.StringAttribute{
				Description: "The category type (different than 'type').",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the version was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The ID of the category.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"order": schema.Int64Attribute{
				Description: "The order of the category.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project": schema.StringAttribute{
				Description: "The ID of the project the category is in.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reference": schema.BoolAttribute{
				Description: "Indicates whether the category is a reference or not.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the category.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the category.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The category type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The 'semver-ish' ReadMe version to create the category under.",
				Optional:    true,
				Computed:    true,
			},
			"version_id": schema.StringAttribute{
				Description: "The version ID the category is associated with.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the API Specification and sets the initial Terraform state.
func (r *categoryResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan.
	var plan categoryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the category.
	createParams := readme.CategoryParams{
		Title: plan.Title.ValueString(),
		Type:  plan.Type.ValueString(),
	}
	var slug string
	var apiResponse *readme.APIResponse
	var err error
	if plan.Version.ValueString() == "" {
		var categoryResponse readme.CategorySaved
		apiResponse, err = r.client.Category.Create(
			&categoryResponse,
			createParams,
			apiRequestOptions(plan.Version),
		)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create category.", clientError(err, apiResponse))
		}
		slug = categoryResponse.Slug
	} else {
		var categoryResponse readme.CategoryVersionSaved
		apiResponse, err = r.client.Category.Create(
			&categoryResponse,
			createParams,
			apiRequestOptions(plan.Version),
		)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create versioned category.", clientError(err, apiResponse))
		}
		slug = categoryResponse.Slug
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the category.
	plan, err = r.get(ctx, slug, plan, apiRequestOptions(plan.Version))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create category.",
			"There was a problem retrieving the category after creation.\n"+err.Error())

		return
	}

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Unable to refresh category after creation.",
			"There was a problem setting the state.")

		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *categoryResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state.
	var state categoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the version using the version ID in the state.
	version := versionClean(ctx, r.client, state.VersionID.ValueString())

	// Get the category metadata.
	state, err := r.get(
		ctx,
		state.Slug.ValueString(),
		state,
		apiRequestOptions(types.StringValue(version)),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read category.", err.Error())

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Unable to refresh category.",
			"There was a problem setting the state.",
		)

		return
	}
}

// Update updates the API Specification and sets the updated Terraform state on success.
func (r *categoryResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan and current state.
	var plan, state categoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the category.
	createParams := readme.CategoryParams{
		Title: plan.Title.ValueString(),
		Type:  plan.Type.ValueString(),
	}
	response, apiResponse, err := r.client.Category.Update(
		state.Slug.ValueString(),
		createParams,
		apiRequestOptions(plan.Version),
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update category.", clientError(err, apiResponse))

		return
	}

	// Get the category.
	plan, err = r.get(ctx, response.Slug, plan, apiRequestOptions(plan.Version))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update category.",
			"There was a problem retrieving the category after update.\n"+err.Error())

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the API Specification and removes the Terraform state on success.
func (r *categoryResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state.
	var state categoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the category.
	_, apiResponse, err := r.client.Category.Delete(
		state.Slug.ValueString(),
		apiRequestOptions(state.Version),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to delete category %s.", state.Slug),
			clientError(err, apiResponse),
		)
	}
}

// ImportState imports an API Specification by ID.
func (r *categoryResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

// get is a helper function for retrieving a category and returning the Terraform resource category model for state.
func (r *categoryResource) get(
	ctx context.Context,
	slug string,
	plan categoryModel,
	options readme.RequestOptions,
) (categoryModel, error) {
	var state categoryModel

	// Get the version from ReadMe.
	response, apiResponse, err := r.client.Category.Get(slug, options)
	if err != nil {
		return state, errors.New(clientError(err, apiResponse))
	}

	state = categoryModel{
		CategoryType: types.StringValue(response.CategoryType),
		CreatedAt:    types.StringValue(response.CreatedAt),
		ID:           types.StringValue(response.ID),
		Order:        types.Int64Value(int64(response.Order)),
		Project:      types.StringValue(response.Project),
		Reference:    types.BoolValue(response.Reference),
		Slug:         types.StringValue(response.Slug),
		Title:        types.StringValue(response.Title),
		Type:         types.StringValue(response.Type),
		Version:      types.StringValue(versionClean(ctx, r.client, response.Version)),
		VersionID:    types.StringValue(response.Version),
	}

	return state, nil
}
