package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RepositoryPermissionsModel struct {
	RetainOnDelete    bool         `tfsdk:"retain_on_delete"`
	Project           string       `tfsdk:"project"`
	Slug              string       `tfsdk:"slug"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       Assignments  `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ RepositoryPermissionInterface = &RepositoryPermissionsModel{}

func (m RepositoryPermissionsModel) getProjectKeyAndSlug(ctx context.Context) (projectKey string, slug string) {
	return m.Project, m.Slug
}

func (m RepositoryPermissionsModel) getAssignment(ctx context.Context) Assignments {
	return m.Assignments
}

func NewRepositoryPermissionsModel(plan RepositoryPermissionsModel, assignmentResult *AssignmentResult) *RepositoryPermissionsModel {
	return &RepositoryPermissionsModel{
		RetainOnDelete:    plan.RetainOnDelete,
		Project:           plan.Project,
		Slug:              plan.Slug,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
