package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type ProjectModel struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	ID                types.Int64  `tfsdk:"id"`
	Key               string       `tfsdk:"key"`
	Name              string       `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       Assignments  `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectModel{}

func (m ProjectModel) getProjectKey(ctx context.Context) string {
	return m.Key
}

func (m ProjectModel) getAssignment(ctx context.Context) Assignments {
	return m.Assignments
}

func NewProjectModel(plan ProjectModel, project *bitbucket.Project, assignmentResult *AssignmentResult) *ProjectModel {
	return &ProjectModel{
		RetainOnDelete:    plan.RetainOnDelete,
		ID:                types.Int64Value(project.ID),
		Key:               plan.Key,
		Name:              plan.Name,
		Description:       plan.Description,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
