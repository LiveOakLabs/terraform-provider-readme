package readme

// Types and functions that provide resource plan modifiers for using Markdown front matter for attribute values.
// Front matter is an alternative way to specify parameters for changelogs, custom pages, and docs.

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
)

// readmeFrontMatter represents the front matter keys available to ReadMe changelogs, custom pages, and docs.
type readmeFrontMatter struct {
	Body          string                `yaml:"body,omitempty"`          // changelogs, custom pages, docs
	Category      string                `yaml:"category"`                // docs
	CategorySlug  string                `yaml:"categorySlug"`            // docs
	Error         readme.DocErrorObject `yaml:"error,omitempty"`         // docs
	Hidden        *bool                 `yaml:"hidden"`                  // changelogs, custom pages, docs
	HTML          string                `yaml:"html,omitempty"`          // custom page
	HTMLMode      *bool                 `yaml:"htmlmode"`                // custom page
	Order         *int                  `yaml:"order"`                   // docs
	ParentDoc     string                `yaml:"parentDoc,omitempty"`     // docs
	ParentDocSlug string                `yaml:"parentDocSlug,omitempty"` // docs
	Title         string                `yaml:"title"`                   // changelogs, custom pages, docs
	Type          string                `yaml:"type,omitempty"`          // changelogs, docs
}

// frontMatterModifier provides plan modifiers that parses Markdown front matter for attribute values.
type frontMatterModifier struct {
	fieldName string
}

// Description returns a plain text description of the modifier's behavior.
func (m frontMatterModifier) Description(ctx context.Context) string {
	return "Reads attribute values from Markdown front matter YAML."
}

// MarkdownDescription returns a markdown formatted description of the modifier's behavior.
func (m frontMatterModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// fromMatterString is a plan modifier function for use on a schema string attribute to set a value from front matter.
func fromMatterString(fieldName string) planmodifier.String {
	return frontMatterModifier{
		fieldName: fieldName,
	}
}

// fromMatterBool is a plan modifier function for use on a schema bool attribute to set a value from front matter.
func fromMatterBool(fieldName string) planmodifier.Bool {
	return frontMatterModifier{
		fieldName: fieldName,
	}
}

// fromMatterInt64 is a plan modifier function for use on a schema int64 attribute to set a value from front matter.
func fromMatterInt64(fieldName string) planmodifier.Int64 {
	return frontMatterModifier{
		fieldName: fieldName,
	}
}

// PlanModifyString sets a string attribute's value from the body Markdown front matter if the attribute is not set and
// a matching attribute is set in front matter.
func (m frontMatterModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	var bodyPlanValue types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyPlanValue)...)

	// If the attribute isn't set, check the body front matter.
	if req.ConfigValue.IsNull() {
		value, diag := valueFromFrontMatter(ctx, bodyPlanValue.ValueString(), m.fieldName)
		if diag != "" {
			resp.Diagnostics.AddError("Error parsing front matter.", diag)

			return
		}
		if value != (reflect.Value{}) {
			tflog.Info(ctx, fmt.Sprintf("%s: setting value from front matter", req.Path))
			resp.PlanValue = types.StringValue(value.Interface().(string))

			return
		}
	}
}

// PlanModifyBool sets a bool attribute's value from the body Markdown front matter if the attribute is not set and a
// matching attribute is set in front matter.
func (m frontMatterModifier) PlanModifyBool(
	ctx context.Context,
	req planmodifier.BoolRequest,
	resp *planmodifier.BoolResponse,
) {
	var bodyPlanValue types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyPlanValue)...)

	// If the attribute isn't set, check the body front matter.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		value, diag := valueFromFrontMatter(ctx, bodyPlanValue.ValueString(), m.fieldName)
		if diag != "" {
			resp.Diagnostics.AddError("Error parsing front matter.", diag)

			return
		}
		if value != (reflect.Value{}) && value.CanInterface() {
			resp.PlanValue = types.BoolValue(value.Elem().Bool())
		}
	}
}

// PlanModifyInt64 sets an int64 attribute's value from the body Markdown front matter if the attribute is not set and
// a matching attribute is set in front matter.
func (m frontMatterModifier) PlanModifyInt64(
	ctx context.Context,
	req planmodifier.Int64Request,
	resp *planmodifier.Int64Response,
) {
	var bodyPlanValue types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyPlanValue)...)

	// If the attribute isn't set, check the body front matter.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		value, diag := valueFromFrontMatter(ctx, bodyPlanValue.ValueString(), m.fieldName)
		if diag != "" {
			resp.Diagnostics.AddError("Error parsing front matter.", diag)

			return
		}
		if value != (reflect.Value{}) && value.CanInterface() {
			resp.PlanValue = types.Int64Value(int64(value.Elem().Int()))
		}
	}
}

// getValue parses the 'body' attribute value for Markdown front matter and returns a specified key's value if it's
// present in the front matter.
//
// The `attribute` parameter is the struct field name representing the YAML front matter key.
//
// This returns a `reflect.Value` to be evaluated as needed based on the type and other conditions.
//
// A string value is provided in place of an error for use with the plugin framework's diagnostics package.
func valueFromFrontMatter(ctx context.Context, body, attribute string) (reflect.Value, string) {
	tflog.Info(ctx, fmt.Sprintf("checking body front matter for attribute '%s'", attribute))

	// Get the FrontMatter from the "body" attribute.
	frontMatter := readmeFrontMatter{}
	_, err := frontmatter.Parse(strings.NewReader(body), &frontMatter)
	if err != nil {
		return reflect.Value{}, err.Error()
	}

	tflog.Debug(ctx, fmt.Sprintf("body front matter=%+v", frontMatter))

	// Get the field value matching the attribute.
	v := reflect.ValueOf(frontMatter)
	field := v.FieldByName(attribute)

	// If the field does not exist, return an empty value.
	if field == (reflect.Value{}) {
		return reflect.Value{}, ""
	}

	// If the field exists and is empty, return an empty value.
	if field.IsZero() {
		tflog.Info(ctx, fmt.Sprintf("no front matter found for attribute %s", attribute))

		return reflect.Value{}, ""
	}

	return field, ""
}
