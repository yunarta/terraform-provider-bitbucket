package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectPermissionsModel struct {
	RetainOnDelete    bool         `tfsdk:"retain_on_delete"`
	Key               string       `tfsdk:"key"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       Assignments  `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectPermissionsModel{}

func (m ProjectPermissionsModel) getProjectKey(ctx context.Context) string {
	return m.Key
}

func (m ProjectPermissionsModel) getAssignment(ctx context.Context) Assignments {
	return m.Assignments
}

func NewProjectPermissionsModel(plan ProjectPermissionsModel, assignmentResult *AssignmentResult) *ProjectPermissionsModel {
	return &ProjectPermissionsModel{
		RetainOnDelete:    plan.RetainOnDelete,
		Key:               plan.Key,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
