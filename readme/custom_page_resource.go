package readme

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &customPageResource{}
	_ resource.ResourceWithConfigure   = &customPageResource{}
	_ resource.ResourceWithImportState = &customPageResource{}
)

// customPageResource is the data source implementation.
type customPageResource struct {
	client *readme.Client
}

// customPageResourceModel is the resource model used by the readme_custom_page resource.
type customPageResourceModel struct {
	Algolia    types.Object `tfsdk:"algolia"`
	Body       types.String `tfsdk:"body"`
	BodyClean  types.String `tfsdk:"body_clean"`
	CreatedAt  types.String `tfsdk:"created_at"`
	FullScreen types.Bool   `tfsdk:"fullscreen"`
	HTML       types.String `tfsdk:"html"`
	HTMLClean  types.String `tfsdk:"html_clean"`
	HTMLMode   types.Bool   `tfsdk:"html_mode"`
	Hidden     types.Bool   `tfsdk:"hidden"`
	ID         types.String `tfsdk:"id"`
	Metadata   types.Object `tfsdk:"metadata"`
	Revision   types.Int64  `tfsdk:"revision"`
	Slug       types.String `tfsdk:"slug"`
	Title      types.String `tfsdk:"title"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// NewCustomPageResource is a helper function to simplify the provider implementation.
func NewCustomPageResource() resource.Resource {
	return &customPageResource{}
}

// Metadata returns the data source type name.
func (r *customPageResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_custom_page"
}

// Configure adds the provider configured client to the data source.
func (r *customPageResource) Configure(
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
func (r customPageResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data customPageResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.Title.IsNull() {
		// check front matter for 'title'.
		titleMatter, diag := valueFromFrontMatter(ctx, data.Body.ValueString(), "Title")
		if diag != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("title"),
				"Error checking front matter during validation.",
				diag,
			)

			return
		}

		// Fail if title is not set in front matter or the attribute.
		if titleMatter == (reflect.Value{}) {
			resp.Diagnostics.AddAttributeError(
				path.Root("title"),
				"Missing required attribute.",
				"'title' must be set using the attribute or in the body front matter.",
			)

			return
		}
	}
}

// Create creates the custom page and sets the initial Terraform state.
func (r *customPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customPageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := readme.CustomPageParams{
		Title:    plan.Title.ValueString(),
		Body:     plan.Body.ValueString(),
		HTML:     plan.HTML.ValueString(),
		HTMLMode: plan.HTMLMode.ValueBoolPointer(),
		Hidden:   plan.Hidden.ValueBoolPointer(),
	}

	page, _, err := r.client.CustomPage.Create(params)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create custom page.", err.Error())

		return
	}

	state := customPageResourceMapToModel(page, plan)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *customPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan, state customPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, _, err := r.client.CustomPage.Get(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve custom page.", err.Error())

		return
	}

	state = customPageResourceMapToModel(page, plan)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the custom page and sets the updated Terraform state on success.
func (r *customPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state customPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := readme.CustomPageParams{
		Title:    plan.Title.ValueString(),
		Body:     plan.Body.ValueString(),
		HTML:     plan.HTML.ValueString(),
		HTMLMode: plan.HTMLMode.ValueBoolPointer(),
		Hidden:   plan.Hidden.ValueBoolPointer(),
	}

	page, _, err := r.client.CustomPage.Update(state.Slug.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update custom page.", err.Error())

		return
	}

	state = customPageResourceMapToModel(page, plan)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the custom page and removes the Terraform state on success.
func (r *customPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, apiResponse, err := r.client.CustomPage.Delete(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete custom page", clientError(err, apiResponse))
	}
}

// ImportState imports a custom page by its slug.
func (r *customPageResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import by slug.
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

// Schema for the readme_custom_page resource.
func (r *customPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage custom pages on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/createcustompage> for more information about this API endpoint.\n\n" +
			"Refer to <https://docs.readme.com/main/docs/rdme> for more information about using front matter in " +
			"ReadMe docs and custom pages.",
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				Description: "The title of the custom page. This can also be set using the `title` front matter.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					fromMatterString("Title"),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the custom page.",
				Computed:    true,
			},
			"body": schema.StringAttribute{
				Description: "The body of the custom page.",
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
			},
			"body_clean": schema.StringAttribute{
				Description: "The body of the custom page after normalization.",
				Computed:    true,
			},
			"html": schema.StringAttribute{
				Description: "The body source formatted in HTML. Only displayed if `htmlmode` is set to `true`. " +
					"Leading and trailing whitespace and certain HTML tags are removed when uploaded to ReadMe. " +
					"The `html_clean` attribute will contain the normalized HTML.",
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString(""),
			},
			"html_clean": schema.StringAttribute{
				Description: "The body formatted in HTML after normalization.",
				Computed:    true,
			},
			"html_mode": schema.BoolAttribute{
				Description: "Set to `true` if `html` should be displayed, otherwise `body` will be displayed.",
				Computed:    true,
				Optional:    true,
			},
			"fullscreen": schema.BoolAttribute{
				Description: "Whether the custom page is in fullscreen mode.",
				Computed:    true,
			},
			"hidden": schema.BoolAttribute{
				Description: "Whether the custom page is hidden.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					fromMatterBool("Hidden"),
				},
			},
			"revision": schema.Int64Attribute{
				Description: "The revision of the custom page.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the custom page.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The date the custom page was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The date the custom page was last updated.",
				Computed:    true,
			},
			"algolia": schema.SingleNestedAttribute{
				Description: "Metadata about the Algolia search integration. " +
					"See <https://docs.readme.com/main/docs/search> for more information.",
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"publish_pending": schema.BoolAttribute{
						Computed: true,
					},
					"record_count": schema.Int64Attribute{
						Computed: true,
					},
					"updated_at": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"metadata": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Computed: true,
					},
					"image": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
					"title": schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}
