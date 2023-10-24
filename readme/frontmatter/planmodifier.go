package frontmatter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// FrontMatterModifier provides plan modifiers that parses Markdown front matter for attribute values.
type FrontMatterModifier struct {
	fieldName string
}

// Description returns a plain text description of the modifier's behavior.
func (m FrontMatterModifier) Description(ctx context.Context) string {
	return "Reads attribute values from Markdown front matter YAML."
}

// MarkdownDescription returns a markdown formatted description of the modifier's behavior.
func (m FrontMatterModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// GetString is a plan modifier function for use on a schema string attribute to set a value from front matter.
func GetString(fieldName string) planmodifier.String {
	return FrontMatterModifier{
		fieldName: fieldName,
	}
}

// GetBool is a plan modifier function for use on a schema bool attribute to set a value from front matter.
func GetBool(fieldName string) planmodifier.Bool {
	return FrontMatterModifier{
		fieldName: fieldName,
	}
}

// GetInt64 is a plan modifier function for use on a schema int64 attribute to set a value from front matter.
func GetInt64(fieldName string) planmodifier.Int64 {
	return FrontMatterModifier{
		fieldName: fieldName,
	}
}

// PlanModifyString sets a string attribute's value from the body Markdown front matter if the attribute is not set and
// a matching attribute is set in front matter.
func (m FrontMatterModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	var bodyPlanValue types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyPlanValue)...)

	// If the attribute isn't set, check the body front matter.
	if req.ConfigValue.IsNull() {
		value, diag := GetValue(ctx, bodyPlanValue.ValueString(), m.fieldName)
		if diag != "" {
			resp.Diagnostics.AddError("Error parsing front matter.", diag)

			return
		}
		if value != (reflect.Value{}) {
			tflog.Debug(ctx, fmt.Sprintf("%s: setting value from front matter", req.Path))
			resp.PlanValue = types.StringValue(value.Interface().(string))

			return
		}
	}
}

// PlanModifyBool sets a bool attribute's value from the body Markdown front matter if the attribute is not set and a
// matching attribute is set in front matter.
func (m FrontMatterModifier) PlanModifyBool(
	ctx context.Context,
	req planmodifier.BoolRequest,
	resp *planmodifier.BoolResponse,
) {
	var bodyPlanValue types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyPlanValue)...)

	// If the attribute isn't set, check the body front matter.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		value, diag := GetValue(ctx, bodyPlanValue.ValueString(), m.fieldName)
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
func (m FrontMatterModifier) PlanModifyInt64(
	ctx context.Context,
	req planmodifier.Int64Request,
	resp *planmodifier.Int64Response,
) {
	var bodyPlanValue types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyPlanValue)...)

	// If the attribute isn't set, check the body front matter.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		value, diag := GetValue(ctx, bodyPlanValue.ValueString(), m.fieldName)
		if diag != "" {
			resp.Diagnostics.AddError("Error parsing front matter.", diag)

			return
		}
		if value != (reflect.Value{}) && value.CanInterface() {
			resp.PlanValue = types.Int64Value(int64(value.Elem().Int()))
		}
	}
}
