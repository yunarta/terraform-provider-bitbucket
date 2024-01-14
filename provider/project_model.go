package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type ProjectModel struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	ID                types.Int64  `tfsdk:"id"`
	Key               types.String `tfsdk:"key"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectModel{}

func (m ProjectModel) getProjectKey(ctx context.Context) string {
	return m.Key.ValueString()
}

func (m ProjectModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := m.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func NewProjectModel(plan ProjectModel, project *bitbucket.Project, assignmentResult *AssignmentResult) *ProjectModel {
	return &ProjectModel{
		RetainOnDelete:    plan.RetainOnDelete,
		ID:                types.Int64Value(project.ID),
		Key:               types.StringValue(project.Key),
		Name:              types.StringValue(project.Name),
		Description:       types.StringValue(project.Description),
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
