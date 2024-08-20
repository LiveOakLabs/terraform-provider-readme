package readme

import (
	"bytes"
	"context"
	"crypto/sha512"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &imageResource{}
	_ resource.ResourceWithConfigure = &imageResource{}
)

// imageResource is the resource implementation.
type imageResource struct {
	client *readme.Client
	config providerConfig
}

// imageResourceModel is the data structure used to hold the resource state.
type imageResourceModel struct {
	Color    types.String `tfsdk:"color"`
	Filename types.String `tfsdk:"filename"`
	Height   types.Int64  `tfsdk:"height"`
	ID       types.String `tfsdk:"id"`
	Source   types.String `tfsdk:"source"`
	URL      types.String `tfsdk:"url"`
	Width    types.Int64  `tfsdk:"width"`
}

// NewImageResource is a helper function to simplify the provider implementation.
func NewImageResource() resource.Resource {
	return &imageResource{}
}

// Metadata returns the data source type name.
func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// Configure adds the provider configured client to the data source.
func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg := req.ProviderData.(*providerData)
	r.client = cfg.client
	r.config = cfg.config
}

// openFile returns the contents of a file as bytes.
func openFile(src string) ([]byte, error) {
	// check if file exists
	if _, err := os.Stat(src); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// open file
	file, err := os.Open(filepath.Clean(src))
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	fileBody := &bytes.Buffer{}
	_, err = io.Copy(fileBody, file)
	if err != nil {
		return nil, fmt.Errorf("error copying file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing file: %w", err)
	}

	// return file contents as bytes
	data := fileBody.Bytes()

	return data, nil
}

// sha256Sum returns the sha256 sum of a byte slice.
func sha256Sum(src []byte) string {
	sha_256 := sha512.New512_256()
	sha_256.Write(src)

	return fmt.Sprintf("%x", sha_256.Sum(nil))
}

// imageShasumModifier is a plan modifier that will set the shasum attribute to unknown if the source image has changed.
type imageShasumModifier struct{}

// imageShasumChanged is a plan modifier that will set the shasum attribute to unknown if the source image has changed.
func imageShasumChanged() planmodifier.String {
	return imageShasumModifier{}
}

// Description returns a description of the plan modifier.
func (m imageShasumModifier) Description(ctx context.Context) string {
	return "Plan modifier that updates the shasum attribute if the source image has changed."
}

// MarkdownDescription returns a description of the plan modifier in Markdown format.
func (m imageShasumModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString is called to modify the plan for a string attribute.
func (m imageShasumModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	var sourcePlanValue, sumStateValue types.String
	// Retrieve the current source value from the plan.
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("source"), &sourcePlanValue)...)

	// Get the shasum (id) attribute from the state.
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &sumStateValue)...)

	// Open the source image file and calculate the sha512 sum.
	sourceData, err := openFile(sourcePlanValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get checksum for image file.", err.Error())

		return
	}
	shaSum := sha256Sum(sourceData)

	// If the source image has changed, set the shasum to unknown.
	if shaSum != sumStateValue.ValueString() {
		resp.PlanValue = types.StringUnknown()
	}
}

// Schema defines the image resource attributes.
func (r *imageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Images on ReadMe.com\n\n" +
			"The images API is not part of the official ReadMe API and therefore not documented or fully featured.\n\n" +
			"Images are not truly stateful - the provider tracks the local source image for changes and will upload " +
			"a new image if the source is changed. The provider makes a HEAD request to the image URL to verify its " +
			"existence. Any change to the source path or checksum will trigger a resource replacement.",
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				Description: "The path to the local image source.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"color": schema.StringAttribute{
				Description: "The color of the image.",
				Computed:    true,
			},
			"filename": schema.StringAttribute{
				Description: "The filename of the image.",
				Computed:    true,
			},
			"height": schema.Int64Attribute{
				Description: "The pixel height of the image.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The sha512sum of the source image.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					imageShasumChanged(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Description: "The URL of the uploaded image.",
				Computed:    true,
			},
			"width": schema.Int64Attribute{
				Description: "The pixel width of the image.",
				Computed:    true,
			},
		},
	}
}

// Create a image and set the initial Terraform state.
func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan imageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Open the image source file.
	imageData, err := openFile(plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source image file.", err.Error())

		return
	}

	// Create the image.
	image, apiResponse, err := r.client.Image.Upload(imageData, plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create image.", clientError(err, apiResponse))

		return
	}

	// Create the plan.
	plan = imageResourceModel{
		Color:    types.StringValue(image.Color),
		Filename: types.StringValue(image.Filename),
		Height:   types.Int64Value(int64(image.Height)),
		ID:       types.StringValue(sha256Sum(imageData)),
		Source:   plan.Source,
		URL:      types.StringValue(image.URL),
		Width:    types.Int64Value(int64(image.Width)),
	}

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read the remote state and refresh the Terraform state with the latest data.
func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var plan, state imageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if image exists.
	res, err := http.Head(state.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read remote image.", err.Error())
	}
	err = res.Body.Close()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read remote image.", err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Remove resource if image does not exist remotely.
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for image resources.
func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete is not supported for image resources.
// This removes the resource from state, but does not delete the image from ReadMe.
func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
