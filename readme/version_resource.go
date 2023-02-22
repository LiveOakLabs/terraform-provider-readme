package readme

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lobliveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &versionResource{}
	_ resource.ResourceWithConfigure   = &versionResource{}
	_ resource.ResourceWithImportState = &versionResource{}
)

// versionResource is the data source implementation.
type versionResource struct {
	client *readme.Client
}

// versionResourceModel maps the struct from the ReadMe client library to Terraform resource attributes.
type versionResourceModel struct {
	Categories   types.List   `tfsdk:"categories"`
	Codename     types.String `tfsdk:"codename"`
	CreatedAt    types.String `tfsdk:"created_at"`
	From         types.String `tfsdk:"from"`
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

// NewVersionResource is a helper function to simplify the provider implementation.
func NewVersionResource() resource.Resource {
	return &versionResource{}
}

// Metadata returns the data source type name.
func (r *versionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

// Configure adds the provider configured client to the data source.
func (r *versionResource) Configure(
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
func (r versionResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data versionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.IsStable.ValueBool() && data.IsHidden.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("is_hidden"),
			"A stable version cannot be hidden.", "is_stable and is_hidden cannot both be true. ",
		)
	}

	if data.IsStable.ValueBool() && data.IsDeprecated.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("is_deprecated"),
			"A stable version cannot be deprecated.",
			"is_stable and is_deprecated cannot both be true. ",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema defines the version resource attributes.
func (r *versionResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages Versions on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getversion> for more information about this API " +
			"endpoint.",
		Attributes: map[string]schema.Attribute{
			"categories": schema.ListAttribute{
				Description: "List of category IDs the version is associated with.",
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"codename": schema.StringAttribute{
				Description: "Dubbed name of version.",
				Computed:    true,
				Optional:    true,
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
			"from": schema.StringAttribute{
				Description: "The version this version is derived from. Note that this is only an attribute used for " +
					"initial creation. The ReadMe API otherwise refers to the 'from' value as an ID tracked in the " +
					"forked_from attribute. When importing a version, the from field will be created after the next " +
					"apply.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"forked_from": schema.StringAttribute{
				Description: "The ID of the version this version is derived from.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The ID of the version.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_beta": schema.BoolAttribute{
				Description: "Toggles if the version is beta or not.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_deprecated": schema.BoolAttribute{
				Description: "Toggles if the version is deprecated or not.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_hidden": schema.BoolAttribute{
				Description: "Toggles if the version is hidden or not. A project's stable version cannot be " +
					"set to hidden.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_stable": schema.BoolAttribute{
				Description: "Toggles if the version is stable. A project can only have a single stable version. " +
					"Changing a stable version to non-stable will trigger a replacement. " +
					"The main 'stable' version for a project cannot be deleted.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplaceIf(
						func(
							ctx context.Context,
							req planmodifier.BoolRequest,
							resp *boolplanmodifier.RequiresReplaceIfFuncResponse,
						) {
							// If changed from true to false, require a replacement.
							if req.StateValue.ValueBool() && !req.PlanValue.ValueBool() {
								resp.RequiresReplace = true
							}
						}, "", ""),
				},
			},
			"project": schema.StringAttribute{
				Description: "The project the version is in.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"release_date": schema.StringAttribute{
				Description: "Timestamp of when the version was released.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The version string, usually a semantic version.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_clean": schema.StringAttribute{
				Description: "A 'clean' version string with certain characters replaced, usually a semantic version.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					changedIfOther(path.Root("version")),
				},
			},
		},
	}
}

// Create a version and set the initial Terraform state.
func (r *versionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan.
	var plan versionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the version.
	plan, err := r.save("create", plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create version.", err.Error())

		return
	}

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read the remote state and refresh the Terraform state with the latest data.
func (r *versionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state.
	var plan, state versionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get version metadata.
	state, err := r.get(plan.VersionClean.ValueString(), plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read version metadata.", err.Error())

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update an existing version and set the updated Terraform state on success.
func (r *versionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan and current state.
	var plan, state versionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the version.
	plan, err := r.save("update", plan, state.VersionClean.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update version.", err.Error())

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete a version and remove the Terraform state on success.
func (r *versionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state.
	var state versionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the version.
	_, apiResponse, err := r.client.Version.Delete(state.VersionClean.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to delete version %s.", state.VersionClean),
			clientError(err, apiResponse),
		)

		return
	}
}

// ImportState imports a version via the 'version_clean' attribute.
func (r *versionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Use the "version_clean" attribute for importing.
	resource.ImportStatePassthroughID(ctx, path.Root("version_clean"), req, resp)
}

// get is a helper function for retrieving a version from ReadMe and mapping the response to the resource model.
// The resource module populated with current remote state is returned.
// A string is returned as an error that will be referenced in a caller function's resource response error.
func (r *versionResource) get(
	version string,
	plan versionResourceModel,
) (versionResourceModel, error) {
	var state versionResourceModel

	// Get the version from ReadMe.
	response, apiResponse, err := r.client.Version.Get(version)
	if err != nil {
		return versionResourceModel{}, errors.New(clientError(err, apiResponse))
	}

	if response.ID == "" {
		return versionResourceModel{}, fmt.Errorf(
			"response is empty when looking up version '%s'",
			version,
		)
	}

	// Map response to model.
	state = versionResourceModel{
		Codename:     types.StringValue(response.Codename),
		CreatedAt:    types.StringValue(response.CreatedAt),
		ID:           types.StringValue(response.ID),
		ForkedFrom:   types.StringValue(response.ForkedFrom),
		From:         plan.From,
		IsBeta:       types.BoolValue(response.IsBeta),
		IsDeprecated: types.BoolValue(response.IsDeprecated),
		IsHidden:     types.BoolValue(response.IsHidden),
		IsStable:     types.BoolValue(response.IsStable),
		Project:      types.StringValue(response.Project),
		ReleaseDate:  types.StringValue(response.ReleaseDate),
		Version:      types.StringValue(response.Version),
		VersionClean: types.StringValue(response.VersionClean),
	}

	// Map category list to Terraform schema model.
	categories := make([]attr.Value, 0, len(response.Categories))
	for _, cat := range response.Categories {
		categories = append(categories, types.StringValue(cat))
	}

	state.Categories, _ = types.ListValue(types.StringType, categories)

	return state, nil
}

// save is a helper function to create or update a version.
// The version is returned as a versionResourceModel.
// A string is returned in the second position for an error message that the caller function references in its
// response.
func (r *versionResource) save(
	action string,
	plan versionResourceModel,
	version ...string,
) (versionResourceModel, error) {
	var createdVersion readme.Version
	var err error
	var apiResponse *readme.APIResponse

	createParams := readme.VersionParams{
		Codename:     plan.Codename.ValueString(),
		From:         plan.From.ValueString(),
		IsBeta:       boolPoint(plan.IsBeta.ValueBool()),
		IsDeprecated: boolPoint(plan.IsDeprecated.ValueBool()),
		IsHidden:     boolPoint(plan.IsHidden.ValueBool()),
		IsStable:     boolPoint(plan.IsStable.ValueBool()),
		Version:      plan.Version.ValueString(),
	}

	if action == "update" {
		createdVersion, apiResponse, err = r.client.Version.Update(version[0], createParams)
	} else {
		createdVersion, apiResponse, err = r.client.Version.Create(createParams)
	}

	if err != nil {
		return versionResourceModel{}, errors.New(clientError(err, apiResponse))
	}

	plan, err = r.get(createdVersion.VersionClean, plan)
	if err != nil {
		return plan, err
	}

	return plan, nil
}
