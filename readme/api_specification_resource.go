package readme

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &apiSpecificationResource{}
	_ resource.ResourceWithConfigure   = &apiSpecificationResource{}
	_ resource.ResourceWithImportState = &apiSpecificationResource{}
)

// apiSpecificationResource is the data source implementation.
type apiSpecificationResource struct {
	client *readme.Client
}

// apiSpecificationResourceModel maps the struct from the ReadMe client library to Terraform attributes.
type apiSpecificationResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Category   types.Object `tfsdk:"category"`
	UUID       types.String `tfsdk:"uuid"`
	Definition types.String `tfsdk:"definition"`
	LastSynced types.String `tfsdk:"last_synced"`
	Semver     types.String `tfsdk:"semver"`
	Source     types.String `tfsdk:"source"`
	Title      types.String `tfsdk:"title"`
	Type       types.String `tfsdk:"type"`
	Version    types.String `tfsdk:"version"`
}

// NewAPISpecificationResource is a helper function to simplify the provider implementation.
func NewAPISpecificationResource() resource.Resource {
	return &apiSpecificationResource{}
}

// Metadata returns the data source type name.
func (r *apiSpecificationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_api_specification"
}

// Configure adds the provider configured client to the data source.
func (r *apiSpecificationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*readme.Client)
}

// jsonMatch compares two JSON strings without regards to formatting and returns a bool.
// This function is used for comparing API specifications.
func jsonMatch(one, two string) (bool, error) {
	var oneIntf, twoIntf interface{}
	err := json.Unmarshal([]byte(one), &oneIntf)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling first item: %w", err)
	}
	err = json.Unmarshal([]byte(two), &twoIntf)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling second item: %w", err)
	}

	if reflect.DeepEqual(oneIntf, twoIntf) {
		return true, nil
	}

	return false, nil
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
func (r *apiSpecificationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages an API specification on ReadMe.com\n\n" +
			"The provider creates and updates API specifications by first uploading the definition to the " +
			"API registry and then creating or updating the API specification using the UUID returned from the " +
			"API registry. This is necessary for associating an API specification with its definition. Ensuring " +
			"the definition is created in the API registry is necessary for retrieving the " +
			"remote definition. This behavior is undocumented in the ReadMe API documentation but works the same way " +
			"the official ReadMe `rdme` CLI tool works.\n\n" +
			"## External Changes\n\n" +
			"External changes made to an API specification managed by Terraform will not be detected due to the way " +
			"the API registry works. When a specification definition is updated, the registry UUID changes and is " +
			"only available from the response when the definition is published to the registry. When Terraform runs " +
			"after an external update, there's no way of programatically retrieving the current state without the " +
			"current UUID. Forcing a Terraform update (e.g. tainting or a manual change) will get things " +
			"synchronized again.\n\n" +
			"## Importing Existing Specifications\n\n" +
			"Importing API specifications is also limited due to the behavior of the API registry and associating a " +
			"specification with its definition. When importing, Terraform will replace the remote definition on its " +
			"next run, regardless if it differs from the local definition. This will associate a registry UUID " +
			"with the specification.\n\n" +
			"See <https://docs.readme.com/main/reference/uploadapispecification> for more information about this API " +
			"endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the API specification.",
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
				Description: "The raw API specification definition JSON.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_synced": schema.StringAttribute{
				Description: "Timestamp of last synchronization.",
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "The API registry UUID associated with the specification.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The creation source of the API specification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the API specification derived from the specification JSON.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the API specification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The version ID the API specification is associated with.",
				Computed:    true,
			},
			"semver": schema.StringAttribute{
				Description: "The semver(-ish) of the API specification. This value may also be set in the " +
					"definition JSON `info:version` key, but will be ignored if this attribute is set. Changing the " +
					"version of a created resource will replace the API specification. Use unique resources to use " +
					"the same specification across multiple versions.\n\n" +
					"Learn more about document versioning at <https://docs.readme.com/main/docs/versions>.",
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
func (r *apiSpecificationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan.
	var plan apiSpecificationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the specification.
	plan, err := r.save("create", "", plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create API specification.", err.Error())

		return
	}

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *apiSpecificationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state.
	var plan, state apiSpecificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Track the current state definition.
	currentDefinition := state.Definition

	// Determine the version ID of the specification from the state or in the plan.
	version := ""
	if plan.Version.ValueString() != "" {
		version = "id:" + plan.Version.ValueString()
	}

	// Get the spec definition from the API registry if a registry UUID is available in the registry.
	// When importing, we do not have a UUID available and set the definition to empty.
	var remoteDefinition types.String
	if state.UUID.ValueString() != "" {
		def, apiResponse, err := r.client.APIRegistry.Get(state.UUID.ValueString())
		if err != nil {
			if apiResponse.APIErrorResponse.Error == "SPEC_NOTFOUND" {
				resp.State.RemoveResource(ctx)

				return
			}

			resp.Diagnostics.AddError(
				"Unable to read API specification.",
				clientError(err, apiResponse),
			)

			return
		}

		remoteDefinition = types.StringValue(def)
	}

	// Get the spec plan.
	state, err := r.makePlan(
		state.ID.ValueString(),
		currentDefinition,
		state.UUID.ValueString(),
		version,
	)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read API specification.", err.Error())

		return
	}

	// Compare the local state with the remote definition.
	// The JSON keys/values are compared between the local and remote definition without regards to whitespace.
	// Only update the state if they truly differ.
	match, _ := jsonMatch(currentDefinition.ValueString(), remoteDefinition.ValueString())
	if !match {
		state.Definition = remoteDefinition
	} else {
		state.Definition = currentDefinition
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the API Specification and sets the updated Terraform state on success.
func (r *apiSpecificationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan and current state.
	var plan, state apiSpecificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the specification.
	plan, err := r.save("update", state.ID.ValueString(), plan)
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
func (r *apiSpecificationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state.
	var state apiSpecificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, apiResponse, err := r.client.APISpecification.Delete(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete API specification.",
			clientError(err, apiResponse),
		)

		return
	}
}

// ImportState imports an API Specification by ID.
func (r *apiSpecificationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Use the "id" attribute for importing.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// save is a helper function that performs the specified action and returns the responses.
//
// The provided plan definition is created in the ReadMe API registry, followed by a create or update of the
// specification itself using the registry UUID.
//
// After creation or update, the specification is retrieved and `makePlan()` is called to map the results to the
// Terraform resource schema that is returned.
func (r *apiSpecificationResource) save(
	action, specID string, plan apiSpecificationResourceModel,
) (apiSpecificationResourceModel, error) {
	var registry readme.APIRegistrySaved
	var response readme.APISpecificationSaved
	var apiResponse *readme.APIResponse

	// If a semver is specified, use that.
	version := ""
	if plan.Semver.ValueString() != "" {
		version = plan.Semver.ValueString()
	}

	// Upload the API specification to the API registry.
	registry, err := r.createRegistry(plan.Definition.ValueString(), version)
	if err != nil {
		return apiSpecificationResourceModel{}, err
	}

	// Create or update an API specification associated with the API registry.
	requestOptions := readme.RequestOptions{Version: version}
	if action == "update" {
		response, apiResponse, err = r.client.APISpecification.Update(
			specID,
			"uuid:"+registry.RegistryUUID,
		)
	} else {
		response, apiResponse, err = r.client.APISpecification.Create(
			"uuid:"+registry.RegistryUUID,
			requestOptions,
		)
	}

	if err != nil {
		return apiSpecificationResourceModel{}, errors.New(clientError(err, apiResponse))
	}

	if response.ID == "" {
		return apiSpecificationResourceModel{}, errors.New("response is empty after saving")
	}

	// Get the spec plan.
	plan, err = r.makePlan(response.ID, plan.Definition, registry.RegistryUUID, version)
	if err != nil {
		return apiSpecificationResourceModel{}, err
	}

	return plan, nil
}

// makePlan is a helper function that responds with a computed Terraform resource schema.
//
// If a version ID is provided instead of a semver, a call to the version API is
// made to determine the semver.
//
// `get()` is called to retrieve the remote specification that is mapped to the schema that is returned.
func (r *apiSpecificationResource) makePlan(
	specID string,
	definition types.String,
	registryUUID, version string,
) (apiSpecificationResourceModel, error) {
	if strings.HasPrefix(version, "id:") {
		versionInfo, _, err := r.client.Version.Get(version)
		if err != nil {
			return apiSpecificationResourceModel{}, fmt.Errorf("%w", err)
		}

		version = versionInfo.VersionClean
	}

	// Retrieve metadata about the API specification.
	spec, err := r.get(specID, version)
	if err != nil {
		return apiSpecificationResourceModel{}, err
	}
	// Map the plan to the resource struct.
	plan := apiSpecificationResourceModel{
		Category:   specCategoryObject(spec),
		Definition: definition,
		ID:         types.StringValue(spec.ID),
		LastSynced: types.StringValue(spec.LastSynced),
		Semver:     types.StringValue(version),
		Source:     types.StringValue(spec.Source),
		Title:      types.StringValue(spec.Title),
		Type:       types.StringValue(spec.Type),
		UUID:       types.StringValue(registryUUID),
		Version:    types.StringValue(spec.Version),
	}

	return plan, nil
}

// get is a helper function that retrieves a specification by ID and returns a readme.APISpecification struct.
func (r *apiSpecificationResource) get(specID, version string) (readme.APISpecification, error) {
	requestOptions := readme.RequestOptions{Version: version}
	specification, apiResponse, err := r.client.APISpecification.Get(specID, requestOptions)
	if err != nil {
		return specification, errors.New(clientError(err, apiResponse))
	}

	if specification.ID == "" {
		return specification, fmt.Errorf("response is empty for specification ID %s", specID)
	}

	return specification, nil
}

// createRegistry is a helper function that creates an API registry definition in ReadMe. This is done before any create
// or update of an API specification.
func (r *apiSpecificationResource) createRegistry(
	definition, version string,
) (readme.APIRegistrySaved, error) {
	registry, apiResponse, err := r.client.APIRegistry.Create(definition, version)
	if err != nil {
		return readme.APIRegistrySaved{}, errors.New(clientError(err, apiResponse))
	}

	return registry, nil
}
