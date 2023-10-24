package otherattributemodifier

// Plan modifier for planning a change for an attribute if another specified
// *bool* attribute changes.

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/terraform-provider-readme/readme/frontmatter"
)

// otherBoolChanged is a plan modifier that plans a change for an
// attribute if another specified *bool* attribute is changed.
type otherBoolChanged struct {
	otherAttribute   path.Path
	otherField       string
	checkFrontmatter bool
	req              interface{}
	resp             interface{}
}

// BoolModifyString is a plan modifier to flag a *string* attribute
// for change if another specified *bool* attribute changes.
// The attribute argument is the path to the attribute to check.
// The otherField argument is the name of the Go struct field in the
// frontmatter.
// The checkFrontmatter argument is a boolean to indicate whether to
// check the frontmatter if the attribute is not set.
func BoolModifyString(
	attribute path.Path,
	otherField string,
	checkFrontmatter bool,
) planmodifier.String {
	return otherBoolChanged{
		otherAttribute:   attribute,
		otherField:       otherField,
		checkFrontmatter: checkFrontmatter,
	}
}

// BoolModifyInt64 is a plan modifier to flag an *int64* attribute for
// change if another specified *bool* attribute changes.
// The attribute argument is the path to the attribute to check.
// The otherField argument is the name of the Go struct field in the
// frontmatter.
// The checkFrontmatter argument is a boolean to indicate whether to
// check the frontmatter if the attribute is not set.
func BoolModifyInt64(
	attribute path.Path,
	otherField string,
	checkFrontmatter bool,
) planmodifier.Int64 {
	return otherBoolChanged{
		otherAttribute:   attribute,
		otherField:       otherField,
		checkFrontmatter: checkFrontmatter,
	}
}

// BoolModifyBool is a plan modifier to flag a *bool* attribute for
// change if another specified *bool* attribute changes.
// The attribute argument is the path to the attribute to check.
// The otherField argument is the name of the Go struct field in the
// frontmatter.
// The checkFrontmatter argument is a boolean to indicate whether to
// check the frontmatter if the attribute is not set.
func BoolModifyBool(
	attribute path.Path,
	otherField string,
	checkFrontmatter bool,
) planmodifier.Bool {
	return otherBoolChanged{
		otherAttribute:   attribute,
		otherField:       otherField,
		checkFrontmatter: checkFrontmatter,
	}
}

// Description returns a plain text description of the modifier's behavior.
func (m otherBoolChanged) Description(ctx context.Context) string {
	return "If another bool attribute is changed, this attribute will be changed."
}

// MarkdownDescription returns a markdown formatted description of the
// modifier's behavior.
func (m otherBoolChanged) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// Load the config, plan, state, and body plan values based on the type.
// The config is loaded to know whether to check frontmatter if the
// attribute is not set. Load the plan and state to compare across values.
// The body plan value is loaded to check frontmatter if the attribute is not set.
func (m otherBoolChanged) loadValues(
	ctx context.Context,
) (types.Bool, types.Bool, types.Bool, types.String, diag.Diagnostics) {
	var bodyValue types.String
	var configValue, planValue, stateValue types.Bool
	var diags diag.Diagnostics

	switch m.req.(type) {
	case planmodifier.StringRequest:
		req := m.req.(planmodifier.StringRequest)
		resp := m.resp.(*planmodifier.StringResponse)
		rDiag := resp.Diagnostics

		rDiag.Append(req.Config.GetAttribute(ctx, m.otherAttribute, &configValue)...)
		rDiag.Append(req.State.GetAttribute(ctx, m.otherAttribute, &stateValue)...)
		rDiag.Append(req.Plan.GetAttribute(ctx, m.otherAttribute, &planValue)...)
		rDiag.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyValue)...)
	case planmodifier.Int64Request:
		req := m.req.(planmodifier.Int64Request)
		resp := m.resp.(*planmodifier.Int64Response)
		rDiag := resp.Diagnostics

		rDiag.Append(req.State.GetAttribute(ctx, m.otherAttribute, &stateValue)...)
		rDiag.Append(req.Config.GetAttribute(ctx, m.otherAttribute, &configValue)...)
		rDiag.Append(req.Plan.GetAttribute(ctx, m.otherAttribute, &planValue)...)
		rDiag.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyValue)...)
	case planmodifier.BoolRequest:
		req := m.req.(planmodifier.BoolRequest)
		resp := m.resp.(*planmodifier.BoolResponse)
		rDiag := resp.Diagnostics

		rDiag.Append(req.State.GetAttribute(ctx, m.otherAttribute, &stateValue)...)
		rDiag.Append(req.Config.GetAttribute(ctx, m.otherAttribute, &configValue)...)
		rDiag.Append(req.Plan.GetAttribute(ctx, m.otherAttribute, &planValue)...)
		rDiag.Append(req.Plan.GetAttribute(ctx, path.Root("body"), &bodyValue)...)
	}

	return configValue, stateValue, planValue, bodyValue, diags
}

// modifyAttribute is a helper function to modify the attribute based on the
// type of the request and response.
//
// The repetition is necessary because the request and response types are
// different for each type.
func (m otherBoolChanged) modifyAttribute(ctx context.Context) {
	var isChanged bool
	var diags diag.Diagnostics

	// Load the config, plan, state, and body plan values based on the type.
	otherConfigValue, otherStateValue, otherPlanValue, bodyPlanValue, diags := m.loadValues(ctx)

	// If the attribute isn't set, check the body front matter.
	if m.checkFrontmatter && otherConfigValue.IsNull() {
		value, errStr := frontmatter.GetValue(ctx, bodyPlanValue.ValueString(), m.otherField)
		if errStr != "" {
			diags.AddError("Error parsing front matter.", errStr)

			return
		}

		// If the value from frontmatter is not empty, compare it to the current state.
		if value != (reflect.Value{}) {
			tflog.Debug(ctx, fmt.Sprintf(
				"%s was found in frontmatter with value %s",
				m.otherAttribute, value))

			// If the value from frontmatter is different from the current
			// plan, mark this attribute as changed.
			isChanged = value.Interface().(bool) != otherPlanValue.ValueBool()
		} else {
			tflog.Debug(ctx, fmt.Sprintf(
				"value for %s was not found in frontmatter",
				m.otherAttribute))
		}
	} else {
		// If the attribute is set, compare it to the current state and ignore
		// the frontmatter.
		tflog.Debug(ctx, "otherBoolChanged: not checking front matter")
		isChanged = otherConfigValue != otherStateValue && !otherStateValue.IsNull()
	}

	tflog.Debug(ctx, fmt.Sprintf(
		"otherBoolChanged: %s otherConfigValue (%s) otherStateValue (%s)",
		m.otherAttribute, otherConfigValue, otherStateValue))

	// If the other attribute is changed, set this attribute unknown to trigger
	// an update. Otherwise, set it to the current plan value.
	switch m.req.(type) {
	case planmodifier.StringRequest:
		resp := m.resp.(*planmodifier.StringResponse)
		req := m.req.(planmodifier.StringRequest)

		if isChanged {
			resp.PlanValue = types.StringUnknown()

			return
		}

		resp.PlanValue = req.PlanValue
	case planmodifier.Int64Request:
		resp := m.resp.(*planmodifier.Int64Response)
		req := m.req.(planmodifier.Int64Request)

		if isChanged {
			resp.PlanValue = types.Int64Unknown()

			return
		}

		resp.PlanValue = req.PlanValue
	case planmodifier.BoolRequest:
		resp := m.resp.(*planmodifier.BoolResponse)
		req := m.req.(planmodifier.BoolRequest)

		if isChanged {
			resp.PlanValue = types.BoolUnknown()

			return
		}

		resp.PlanValue = req.PlanValue
	default:
		tflog.Info(ctx, fmt.Sprintf(
			"otherBoolChanged: unknown request type %T", m.req))
	}
}

// PlanModifyString implements a modifier for planning a change for an
// attribute if another specified *bool* attribute changes.
func (m otherBoolChanged) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	m.modifyAttribute(ctx)
}

// PlanModifyInt64 implements a modifier for planning a change for an
// *int64* attribute if another specified *bool* attribute changes.
// The string attribute is optionally cheked for a value in the frontmatter.
func (m otherBoolChanged) PlanModifyInt64(
	ctx context.Context,
	req planmodifier.Int64Request,
	resp *planmodifier.Int64Response,
) {
	m.modifyAttribute(ctx)
}

// PlanModifyBool implements a modifier for planning a change for a *bool*
// attribute if another specified *bool* attribute changes.
func (m otherBoolChanged) PlanModifyBool(
	ctx context.Context,
	req planmodifier.BoolRequest,
	resp *planmodifier.BoolResponse,
) {
	m.modifyAttribute(ctx)
}
