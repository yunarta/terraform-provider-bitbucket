package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectPermissionsModel0 struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	Key               types.String `tfsdk:"key"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

type ProjectPermissionsModel struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	Project           types.String `tfsdk:"project"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectPermissionsModel{}

func (m ProjectPermissionsModel) getProjectKey(ctx context.Context) string {
	return m.Project.ValueString()
}

func (m ProjectPermissionsModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := m.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func FromProjectPermissionsModel0(plan ProjectPermissionsModel0) *ProjectPermissionsModel {
	return &ProjectPermissionsModel{
		RetainOnDelete:    types.BoolValue(plan.RetainOnDelete.ValueBool()),
		Project:           plan.Key,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     plan.ComputedUsers,
		ComputedGroups:    plan.ComputedGroups,
	}
}

func NewProjectPermissionsModel(plan ProjectPermissionsModel, assignmentResult *AssignmentResult) *ProjectPermissionsModel {
	return &ProjectPermissionsModel{
		RetainOnDelete:    types.BoolValue(plan.RetainOnDelete.ValueBool()),
		Project:           plan.Project,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
