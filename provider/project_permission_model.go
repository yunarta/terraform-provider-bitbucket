package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectPermissionsModel struct {
	Key               types.String `tfsdk:"key"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectPermissionsModel{}

func (m ProjectPermissionsModel) getProjectKey(ctx context.Context) string {
	return m.Key.ValueString()
}

func (m ProjectPermissionsModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := m.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func NewProjectPermissionsModel(plan ProjectPermissionsModel, assignmentResult *AssignmentResult) *ProjectPermissionsModel {
	return &ProjectPermissionsModel{
		Key:               plan.Key,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
