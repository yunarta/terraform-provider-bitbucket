package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RepositoryPermissionsModel struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	Project           types.String `tfsdk:"project"`
	Slug              types.String `tfsdk:"slug"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ RepositoryPermissionInterface = &RepositoryPermissionsModel{}

func (m RepositoryPermissionsModel) getProjectKeyAndSlug(ctx context.Context) (projectKey string, slug string) {
	return m.Project.ValueString(), m.Slug.ValueString()
}

func (m RepositoryPermissionsModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := m.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func NewRepositoryPermissionsModel(plan RepositoryPermissionsModel, assignmentResult *AssignmentResult) *RepositoryPermissionsModel {
	return &RepositoryPermissionsModel{
		RetainOnDelete:    types.BoolValue(plan.RetainOnDelete.ValueBool()),
		Project:           plan.Project,
		Slug:              plan.Slug,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
