package readme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &docSearchDataSource{}
	_ datasource.DataSourceWithConfigure = &docSearchDataSource{}
)

// docSearchDataSource is the data source implementation.
type docSearchDataSource struct {
	client *readme.Client
}

// docSearchResultsModel represents the results from the API when searching for docs.
type docSearchResultsModel struct {
	ID      types.String      `tfsdk:"id"`
	Query   types.String      `tfsdk:"query"`
	Results *[]docSearchModel `tfsdk:"results"`
}

// docSearchModel represents a doc item returned by a search query.
type docSearchModel struct {
	HighlightResult docSearchModelHighlightResult `tfsdk:"highlight_result"`
	IndexName       types.String                  `tfsdk:"index_name"`
	InternalLink    types.String                  `tfsdk:"internal_link"`
	IsReference     types.Bool                    `tfsdk:"is_reference"`
	LinkURL         types.String                  `tfsdk:"link_url"`
	Method          types.String                  `tfsdk:"method"`
	ObjectID        types.String                  `tfsdk:"object_id"`
	Project         types.String                  `tfsdk:"project"`
	ReferenceID     types.String                  `tfsdk:"reference_id"`
	Slug            types.String                  `tfsdk:"slug"`
	SnippetResult   docSearchModelSnippetResult   `tfsdk:"snippet_result"`
	Subdomain       types.String                  `tfsdk:"subdomain"`
	Title           types.String                  `tfsdk:"title"`
	Type            types.String                  `tfsdk:"type"`
	URL             types.String                  `tfsdk:"url"`
	Version         types.String                  `tfsdk:"version"`
}

// docSearchModelHighlightResult represents the `highlight_result` object.
type docSearchModelHighlightResult struct {
	Title   docSearchModelHighlightResultValue `tfsdk:"title"`
	Excerpt docSearchModelHighlightResultValue `tfsdk:"excerpt"`
	Body    docSearchModelHighlightResultValue `tfsdk:"body"`
}

// docSearchModelHighlightResultValue represents the `highlight_result` value objects.
type docSearchModelHighlightResultValue struct {
	Value        types.String   `tfsdk:"value"`
	MatchLevel   types.String   `tfsdk:"match_level"`
	MatchedWords []types.String `tfsdk:"matched_words"`
}

// docSearchModelSnippetResult represents the `snippet_result` object.
type docSearchModelSnippetResult struct {
	Title   docSearchModelSnippetResultValue `tfsdk:"title"`
	Excerpt docSearchModelSnippetResultValue `tfsdk:"excerpt"`
	Body    docSearchModelSnippetResultValue `tfsdk:"body"`
}

// docSearchModelSnippetResultValue represents the `snippet_result` value objects.
type docSearchModelSnippetResultValue struct {
	Value      types.String `tfsdk:"value"`
	MatchLevel types.String `tfsdk:"match_level"`
}

// NewDocSearchDataSource is a helper function to simplify the provider implementation.
func NewDocSearchDataSource() datasource.DataSource {
	return &docSearchDataSource{}
}

// Metadata returns the data source type name.
func (d *docSearchDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_doc_search"
}

// docSearchMatchedWords is a helper function that returns a list of matched words in a doc search mapped to the
// appropriate Terraform types.
func docSearchMatchedWords(list []string) []types.String {
	result := []types.String{}
	for _, word := range list {
		result = append(result, types.StringValue(word))
	}

	return result
}

// Read refreshes the Terraform state with the latest data.
func (d *docSearchDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state docSearchResultsModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get doc metadata from ReadMe API
	doc, apiResponse, err := d.client.Doc.Search(state.Query.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve doc metadata.", clientError(err, apiResponse))

		return
	}

	// Map response body to model
	results := []docSearchModel{}
	for _, result := range doc {
		results = append(results, docSearchModel{
			HighlightResult: docSearchModelHighlightResult{
				Body: docSearchModelHighlightResultValue{
					Value:        types.StringValue(result.HighlightResult.Body.Value),
					MatchLevel:   types.StringValue(result.HighlightResult.Body.MatchLevel),
					MatchedWords: docSearchMatchedWords(result.HighlightResult.Body.MatchedWords),
				},
				Excerpt: docSearchModelHighlightResultValue{
					Value:      types.StringValue(result.HighlightResult.Excerpt.Value),
					MatchLevel: types.StringValue(result.HighlightResult.Excerpt.MatchLevel),
					MatchedWords: docSearchMatchedWords(
						result.HighlightResult.Excerpt.MatchedWords,
					),
				},
				Title: docSearchModelHighlightResultValue{
					Value:        types.StringValue(result.HighlightResult.Title.Value),
					MatchLevel:   types.StringValue(result.HighlightResult.Title.MatchLevel),
					MatchedWords: docSearchMatchedWords(result.HighlightResult.Title.MatchedWords),
				},
			},
			IndexName:    types.StringValue(result.IndexName),
			InternalLink: types.StringValue(result.InternalLink),
			IsReference:  types.BoolValue(result.IsReference),
			LinkURL:      types.StringValue(result.LinkURL),
			Method:       types.StringValue(result.Method),
			ObjectID:     types.StringValue(result.ObjectID),
			Project:      types.StringValue(result.Project),
			ReferenceID:  types.StringValue(result.ReferenceID),
			Slug:         types.StringValue(result.Slug),
			SnippetResult: docSearchModelSnippetResult{
				Body: docSearchModelSnippetResultValue{
					Value:      types.StringValue(result.SnippetResult.Body.Value),
					MatchLevel: types.StringValue(result.SnippetResult.Body.MatchLevel),
				},
				Excerpt: docSearchModelSnippetResultValue{
					Value:      types.StringValue(result.SnippetResult.Excerpt.Value),
					MatchLevel: types.StringValue(result.SnippetResult.Excerpt.MatchLevel),
				},
				Title: docSearchModelSnippetResultValue{
					Value:      types.StringValue(result.SnippetResult.Title.Value),
					MatchLevel: types.StringValue(result.SnippetResult.Title.MatchLevel),
				},
			},
			Subdomain: types.StringValue(result.Subdomain),
			Title:     types.StringValue(result.Title),
			Type:      types.StringValue(result.Type),
			URL:       types.StringValue(result.URL),
			Version:   types.StringValue(result.Version),
		})
	}

	query := state.Query
	state = docSearchResultsModel{
		Query:   query,
		Results: &results,
	}

	// The ID isn't returned in the data source but is tracked internally and required for testing.
	// See https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	state.ID = types.StringValue("readme_doc_search")

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *docSearchDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}

// Schema for the readme_doc_search data source.
func (d *docSearchDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve docs matching a search query on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getdoc> for more information about this API endpoint.\n\n",
		Attributes: map[string]schema.Attribute{
			// The 'id' isn't returned by ReadMe - it's for Terraform use to track state.
			// See https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
			"id": schema.StringAttribute{
				Description: "The internal ID of this resource.",
				Computed:    true,
			},
			"query": schema.StringAttribute{
				Required: true,
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"highlight_result": schema.SingleNestedAttribute{
							Description: "",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"title": schema.SingleNestedAttribute{
									Description: "",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Computed: true,
										},
										"match_level": schema.StringAttribute{
											Computed: true,
										},
										"matched_words": schema.ListAttribute{
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
								"excerpt": schema.SingleNestedAttribute{
									Description: "",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Computed: true,
										},
										"match_level": schema.StringAttribute{
											Computed: true,
										},
										"matched_words": schema.ListAttribute{
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
								"body": schema.SingleNestedAttribute{
									Description: "",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Computed: true,
										},
										"match_level": schema.StringAttribute{
											Computed: true,
										},
										"matched_words": schema.ListAttribute{
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
							},
						},
						"index_name": schema.StringAttribute{
							Computed: true,
						},
						"internal_link": schema.StringAttribute{
							Computed: true,
						},
						"is_reference": schema.BoolAttribute{
							Computed: true,
						},
						"link_url": schema.StringAttribute{
							Computed: true,
						},
						"method": schema.StringAttribute{
							Computed: true,
						},
						"object_id": schema.StringAttribute{
							Computed: true,
						},
						"project": schema.StringAttribute{
							Computed: true,
						},
						"reference_id": schema.StringAttribute{
							Computed: true,
						},
						"slug": schema.StringAttribute{
							Computed: true,
						},
						"snippet_result": schema.SingleNestedAttribute{
							Description: "",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"title": schema.SingleNestedAttribute{
									Description: "",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Computed: true,
										},
										"match_level": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"excerpt": schema.SingleNestedAttribute{
									Description: "",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Computed: true,
										},
										"match_level": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"body": schema.SingleNestedAttribute{
									Description: "",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"value": schema.StringAttribute{
											Computed: true,
										},
										"match_level": schema.StringAttribute{
											Computed: true,
										},
									},
								},
							},
						},
						"subdomain": schema.StringAttribute{
							Computed: true,
						},
						"title": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"url": schema.StringAttribute{
							Computed: true,
						},
						"version": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}
