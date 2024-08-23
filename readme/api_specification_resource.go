package readme

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

const apiSpecResourceDesc = `
Manages API specifications on ReadMe.com by uploading the definition to the API registry and associating it with the
specification using the returned UUID. This association is necessary for managing the API specification and its
definition. The behavior is similar to the official rdme CLI but is undocumented in the ReadMe API.

## External Changes
External changes to API specifications managed by Terraform are not automatically detected. The UUID changes when a
definition is updated, and the new UUID is only available when published to the registry. To synchronize, force an
update via Terraform (e.g., taint or manual change).

## Importing Existing Specifications
Importing is limited due to how the API registry associates specifications with definitions. Terraform will overwrite
the remote definition on the next run, replacing the UUID.

## Managing Documentation
API specifications on ReadMe automatically create a documentation page, but it isn't managed by Terraform. Use the
readme_doc resource with use_slug to manage the documentation page.

See the ReadMe API documentation at https://docs.readme.com/main/reference/uploadapispecification for more information.
`

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &apiSpecResource{}
	_ resource.ResourceWithConfigure   = &apiSpecResource{}
	_ resource.ResourceWithImportState = &apiSpecResource{}
)

// apiSpecResource is the resource implementation.
type apiSpecResource struct {
	client *readme.Client
	config providerConfig
}

// apiSpecResourceModel maps the struct from the ReadMe client library to Terraform attributes.
type apiSpecResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Category       types.Object `tfsdk:"category"`
	DeleteCategory types.Bool   `tfsdk:"delete_category"`
	UUID           types.String `tfsdk:"uuid"`
	Definition     types.String `tfsdk:"definition"`
	LastSynced     types.String `tfsdk:"last_synced"`
	Semver         types.String `tfsdk:"semver"`
	Source         types.String `tfsdk:"source"`
	Title          types.String `tfsdk:"title"`
	Type           types.String `tfsdk:"type"`
	Version        types.String `tfsdk:"version"`
}

// NewAPISpecificationResource is a helper function to simplify the provider implementation.
func NewAPISpecificationResource() resource.Resource {
	return &apiSpecResource{}
}

// Metadata returns the data source type name.
func (r *apiSpecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_specification"
}

// Configure adds the provider configured client to the data source.
func (r *apiSpecResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg := req.ProviderData.(*providerData)
	r.client = cfg.client
	r.config = cfg.config
}

// specCategoryObject maps a readme.CategorySummary type to a generic ObjectValue and returns the ObjectValue for use
// with the Terraform resource schema.
func specCategoryObject(spec readme.APISpecification) basetypes.ObjectValue {
	object, _ := types.ObjectValue(
		map[string]attr.Type{
			"id":    types.StringType,
			"title": types.StringType,
			"slug":  types.StringType,
			"order": types.Int64Type,
			"type":  types.StringType,
		},
		map[string]attr.Value{
			"id":    types.StringValue(spec.Category.ID),
			"title": types.StringValue(spec.Category.Title),
			"slug":  types.StringValue(spec.Category.Slug),
			"order": types.Int64Value(int64(spec.Category.Order)),
			"type":  types.StringValue(spec.Category.Type),
		})

	return object
}

// Schema defines the API Specification resource attributes.
func (r *apiSpecResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: apiSpecResourceDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the API specification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"category": schema.ObjectAttribute{
				Description: "Category metadata for the API specification.",
				Computed:    true,
				AttributeTypes: map[string]attr.Type{
					"id":    types.StringType,
					"slug":  types.StringType,
					"order": types.Int64Type,
					"title": types.StringType,
					"type":  types.StringType,
				},
			},
			"definition": schema.StringAttribute{
				Description: "Raw API specification definition in JSON format.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"delete_category": schema.BoolAttribute{
				Description: "Delete the associated category when the resource is deleted.",
				Optional:    true,
			},
			"last_synced": schema.StringAttribute{
				Description: "Timestamp of the last synchronization.",
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "UUID of the API registry associated with this specification.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Creation source of the API specification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "Title derived from the specification JSON.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the API specification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "Version ID associated with the API specification.",
				Computed:    true,
			},
			"semver": schema.StringAttribute{
				Description: "Semver (or similar) for the API specification. This value can be set in the `info:version` key " +
					"of the definition JSON, but this parameter takes precedence. Changing the version will replace the API " +
					"specification. Use unique resources for multiple versions. Learn more about document versioning " +
					"[here](https://docs.readme.com/main/docs/versions).",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

// Create creates the API Specification and sets the initial Terraform state.
func (r *apiSpecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan.
	var plan apiSpecResourceModel
	if diags := req.Plan.Get(ctx, &plan); diags.HasError() {
		resp.Diagnostics.Append(diags...)

		return
	}

	// Create the API specification.
	createdPlan, err := r.save(saveParams{
		ctx:    ctx,
		action: saveActionCreate,
		specID: "",
		plan:   plan,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create API specification",
			err.Error(),
		)

		return
	}

	// Set the Terraform state with the created plan.
	if diags := resp.State.Set(ctx, createdPlan); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *apiSpecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state apiSpecResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the version ID of the specification from the state or in the plan.
	version := ""
	if v := state.Version.ValueString(); v != "" {
		version = IDPrefix + v
	}

	// Get the spec definition from the API registry if a registry UUID is available in the registry.
	// When importing, we do not have a UUID available and set the definition to empty.
	var remoteDefinition types.String
	if state.UUID.ValueString() != "" {
		def, apiResponse, err := r.client.APIRegistry.Get(state.UUID.ValueString())
		if err != nil {
			if apiResponse != nil && apiResponse.APIErrorResponse.Error == "SPEC_NOTFOUND" {
				resp.State.RemoveResource(ctx)

				return
			}
			resp.Diagnostics.AddError(
				"Unable to read API specification",
				clientError(err, apiResponse),
			)

			return
		}
		remoteDefinition = types.StringValue(def)
	}

	delCatetory := state.DeleteCategory

	// Generate the spec plan.
	state, err := r.makePlan(makePlanParams{
		ctx:          ctx,
		specID:       state.ID.ValueString(),
		definition:   state.Definition,
		registryUUID: state.UUID.ValueString(),
		version:      version,
	})
	if err != nil {
		if strings.Contains(err.Error(), "API specification not found") || strings.Contains(err.Error(), "no match for version ID") {
			tflog.Warn(ctx, fmt.Sprintf("API specification %s not found. Removing from state.", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)

			return
		}
		resp.Diagnostics.AddError("Unable to read API specification", err.Error())

		return
	}

	state.DeleteCategory = delCatetory

	// Compare the local state with the remote definition and update if they differ.
	if match, _ := jsonMatch(state.Definition.ValueString(), remoteDefinition.ValueString()); !match {
		state.Definition = remoteDefinition
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the API Specification and sets the updated Terraform state on success.
func (r *apiSpecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan and current state.
	var plan, state apiSpecResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the specification.
	plan, err := r.save(saveParams{
		ctx:    ctx,
		action: saveActionUpdate,
		specID: state.ID.ValueString(),
		plan:   plan,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update API specification.", err.Error())

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the API Specification and removes the Terraform state on success.
func (r *apiSpecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state apiSpecResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the API Specification.
	if _, apiResponse, err := r.client.APISpecification.Delete(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete API specification",
			clientError(err, apiResponse),
		)

		return
	}

	// Remove the category if delete_category is set to true.
	if state.DeleteCategory.ValueBool() {
		r.deleteCategory(ctx, state, resp)
	}
}

// deleteCategory removes the associated category if it exists.
func (r *apiSpecResource) deleteCategory(
	ctx context.Context,
	state apiSpecResourceModel,
	resp *resource.DeleteResponse,
) {
	// Extract and clean the category slug.
	catSlug := strings.ReplaceAll(state.Category.Attributes()["slug"].String(), "\"", "")

	// Clean the version ID from the state.
	versionID := state.Version.ValueString()
	version := versionClean(ctx, r.client, versionID)

	// Delete the category.
	opts := readme.RequestOptions{Version: version}
	if _, apiResponse, err := r.client.Category.Delete(catSlug, opts); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete category",
			clientError(err, apiResponse),
		)

		return
	}
}

// ImportState imports an API Specification by ID.
func (r *apiSpecResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Use the "id" attribute for importing.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// jsonMatch compares two JSON strings without regards to formatting and returns a bool.
// This function is used for comparing API specifications.
// jsonMatch compares two JSON strings without regard to formatting and returns a bool.
// This function is used for comparing API specifications.
func jsonMatch(one, two string) (bool, error) {
	if !json.Valid([]byte(one)) {
		return false, fmt.Errorf("invalid JSON in first input: %s", one)
	}
	if !json.Valid([]byte(two)) {
		return false, fmt.Errorf("invalid JSON in second input: %s", two)
	}

	var oneMap, twoMap map[string]interface{}
	if err := json.Unmarshal([]byte(one), &oneMap); err != nil {
		return false, fmt.Errorf("error unmarshalling first JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(two), &twoMap); err != nil {
		// nolint:errorlint
		return false, fmt.Errorf("error unmarshalling second JSON: %w", err)
	}

	return reflect.DeepEqual(oneMap, twoMap), nil
}

type saveParams struct {
	ctx    context.Context
	action saveAction
	specID string
	plan   apiSpecResourceModel
}

// save is a helper function that performs the specified action and returns the responses.
//
// The provided plan definition is created in the ReadMe API registry, followed by a create or update of the
// specification itself using the registry UUID.
//
// After creation or update, the specification is retrieved and `makePlan()` is called to map the results to the
// Terraform resource schema that is returned.
func (r *apiSpecResource) save(params saveParams) (apiSpecResourceModel, error) {
	// Determine the version, preferring semver if specified.
	version := params.plan.Semver.ValueString()

	// Upload the API specification to the registry.
	registry, err := r.createRegistry(params.plan.Definition.ValueString(), version)
	if err != nil {
		return apiSpecResourceModel{}, fmt.Errorf("unable to create registry: %w", err)
	}

	// Prepare request options and perform the save action (create or update).
	requestOptions := readme.RequestOptions{Version: version}
	var response readme.APISpecificationSaved
	var apiResponse *readme.APIResponse

	switch params.action {
	case saveActionUpdate:
		response, apiResponse, err = r.client.APISpecification.Update(params.specID, UUIDPrefix+registry.RegistryUUID)
	case saveActionCreate:
		response, apiResponse, err = r.client.APISpecification.Create(UUIDPrefix+registry.RegistryUUID, requestOptions)
	default:
		return apiSpecResourceModel{}, fmt.Errorf("unknown save action: %v", params.action)
	}

	if err != nil {
		status := 0
		if apiResponse != nil {
			status = apiResponse.HTTPResponse.StatusCode
		}

		return apiSpecResourceModel{}, fmt.Errorf("unable to save specification: (%d) %w", status, err)
	}

	if response.ID == "" {
		return apiSpecResourceModel{}, fmt.Errorf("specification response is empty after saving: %+v", response)
	}

	// Preserve DeleteCategory value and update the plan.
	delCategory := params.plan.DeleteCategory
	plan, err := r.makePlan(makePlanParams{
		ctx:          params.ctx,
		specID:       response.ID,
		definition:   params.plan.Definition,
		registryUUID: registry.RegistryUUID,
		version:      version,
	})
	if err != nil {
		return apiSpecResourceModel{}, fmt.Errorf("unable to make plan: %w", err)
	}
	plan.DeleteCategory = delCategory

	return plan, nil
}

type makePlanParams struct {
	ctx          context.Context
	specID       string
	definition   types.String
	registryUUID string
	version      string
}

// makePlan is a helper function that responds with a computed Terraform resource schema.
//
// If a version ID is provided instead of a semver, a call to the version API is
// made to determine the semver.
// `get()` is called to retrieve the remote specification that is mapped to the schema that is returned.
func (r *apiSpecResource) makePlan(params makePlanParams) (apiSpecResourceModel, error) {
	// Resolve the version if it's an ID.
	if strings.HasPrefix(params.version, IDPrefix) {
		versionInfo, _, err := r.client.Version.Get(params.version)
		if err != nil {
			return apiSpecResourceModel{}, fmt.Errorf("error resolving version: %w", err)
		}
		params.version = versionInfo.VersionClean
	}

	// Retrieve the specification metadata.
	spec, err := r.get(params.ctx, params.specID, params.version)
	if err != nil {
		return apiSpecResourceModel{}, fmt.Errorf("error getting specification: %w", err)
	}

	// Map the retrieved data to the resource model.
	return apiSpecResourceModel{
		Category:   specCategoryObject(spec),
		Definition: params.definition,
		ID:         types.StringValue(spec.ID),
		LastSynced: types.StringValue(spec.LastSynced),
		Semver:     types.StringValue(params.version),
		Source:     types.StringValue(spec.Source),
		Title:      types.StringValue(spec.Title),
		Type:       types.StringValue(spec.Type),
		UUID:       types.StringValue(params.registryUUID),
		Version:    types.StringValue(spec.Version),
	}, nil
}

// get is a helper function that retrieves a specification by ID and returns a readme.APISpecification struct.
func (r *apiSpecResource) get(ctx context.Context, specID, version string) (readme.APISpecification, error) {
	requestOptions := readme.RequestOptions{Version: version}
	specification, _, err := r.client.APISpecification.Get(specID, requestOptions)
	if err != nil {
		return readme.APISpecification{}, fmt.Errorf("unable to get specification ID %s: %w", specID, err)
	}

	if specification.ID == "" {
		return readme.APISpecification{}, fmt.Errorf("specification response is empty for specification ID %s", specID)
	}

	return specification, nil
}

// createRegistry is a helper function that creates an API registry definition in ReadMe.
// This is done before any create or update of an API specification.
func (r *apiSpecResource) createRegistry(definition, version string) (readme.APIRegistrySaved, error) {
	registry, apiResponse, err := r.client.APIRegistry.Create(definition, version)
	if err != nil {
		return readme.APIRegistrySaved{}, fmt.Errorf("unable to create API registry: %s", clientError(err, apiResponse))
	}

	return registry, nil
}
