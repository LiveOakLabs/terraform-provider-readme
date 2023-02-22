package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &categoryDocsDataSource{}
	_ datasource.DataSourceWithConfigure = &categoryDocsDataSource{}
)

// categoryDocsDataSource is the data source implementation.
type categoryDocsDataSource struct {
	client *readme.Client
}

// categoryDocsModel maps the response to the Terraform data source schema.
type categoryDocsModel struct {
	ID   types.String  `tfsdk:"id"`
	Slug types.String  `tfsdk:"slug"`
	Docs []categoryDoc `tfsdk:"docs"`
}

// categoryDocs represents a document within a category.
type categoryDoc struct {
	ID       types.String       `tfsdk:"id"`
	Title    types.String       `tfsdk:"title"`
	Slug     types.String       `tfsdk:"slug"`
	Order    types.Int64        `tfsdk:"order"`
	Hidden   types.Bool         `tfsdk:"hidden"`
	Children []categoryDocChild `tfsdk:"children"`
}

// categoryDocsChild represents a child document within a category.
type categoryDocChild struct {
	ID     types.String `tfsdk:"id"`
	Title  types.String `tfsdk:"title"`
	Slug   types.String `tfsdk:"slug"`
	Order  types.Int64  `tfsdk:"order"`
	Hidden types.Bool   `tfsdk:"hidden"`
}

// NewCategoryDocsDataSource is a helper function to simplify the provider implementation.
func NewCategoryDocsDataSource() datasource.DataSource {
	return &categoryDocsDataSource{}
}

// Metadata returns the data source type name.
func (d *categoryDocsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_category_docs"
}

// Schema defines the schema for the data source.
func (d *categoryDocsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve all docs for a project category on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getcategorydocs> for more information about this API endpoint.\n\n",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The internal Terraform ID of the data source.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The category slug to retrieve docs for.",
				Required:    true,
			},
			"docs": schema.ListNestedAttribute{
				Description: "List of category summaries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"slug": schema.StringAttribute{
							Description: "The slug of the category.",
							Computed:    true,
						},
						"title": schema.StringAttribute{
							Description: "The title of the category.",
							Computed:    true,
						},
						"order": schema.Int64Attribute{
							Description: "The order of the category.",
							Computed:    true,
						},
						"hidden": schema.BoolAttribute{
							Description: "Tye type of category.",
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "The unique ID of the category.",
							Computed:    true,
						},
						"children": schema.ListNestedAttribute{
							Description: "",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"slug": schema.StringAttribute{
										Description: "The slug of the category.",
										Computed:    true,
									},
									"title": schema.StringAttribute{
										Description: "The title of the category.",
										Computed:    true,
									},
									"order": schema.Int64Attribute{
										Description: "The order of the category.",
										Computed:    true,
									},
									"hidden": schema.BoolAttribute{
										Description: "Tye type of category.",
										Computed:    true,
									},
									"id": schema.StringAttribute{
										Description: "The unique ID of the category.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *categoryDocsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state categoryDocsModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get categoryDocs metadata from ReadMe API.
	categoryDocs, apiResponse, err := d.client.Category.GetDocs(state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve category docs.",
			clientError(err, apiResponse),
		)

		return
	}

	for _, catDoc := range categoryDocs {
		// Map response body to model.
		doc := categoryDoc{
			ID:       types.StringValue(catDoc.ID),
			Order:    types.Int64Value(int64(catDoc.Order)),
			Slug:     types.StringValue(catDoc.Slug),
			Title:    types.StringValue(catDoc.Title),
			Hidden:   types.BoolValue(catDoc.Hidden),
			Children: []categoryDocChild{},
		}

		// Map children doc list to Terraform type.
		for _, childDoc := range catDoc.Children {
			child := categoryDocChild{
				ID:     types.StringValue(childDoc.ID),
				Order:  types.Int64Value(int64(childDoc.Order)),
				Slug:   types.StringValue(childDoc.Slug),
				Title:  types.StringValue(childDoc.Title),
				Hidden: types.BoolValue(childDoc.Hidden),
			}
			doc.Children = append(doc.Children, child)
		}

		state.Docs = append(state.Docs, doc)
	}

	state.ID = types.StringValue("readme_category_docs")

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *categoryDocsDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}
