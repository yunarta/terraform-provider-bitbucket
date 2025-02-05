package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
)

type ProjectModel0 struct {
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

type ProjectModel struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	ID                types.Int64  `tfsdk:"id"`
	Project           types.String `tfsdk:"project"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectModel{}

func (m ProjectModel) getProjectKey(ctx context.Context) string {
	return m.Project.ValueString()
}

func (m ProjectModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := m.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func FromProjectModel0(plan ProjectModel0) *ProjectModel {
	return &ProjectModel{
		RetainOnDelete:    plan.RetainOnDelete,
		ID:                plan.ID,
		Project:           plan.Key,
		Name:              plan.Name,
		Description:       plan.Description,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     plan.ComputedUsers,
		ComputedGroups:    plan.ComputedGroups,
	}
}

func NewProjectModel(plan ProjectModel, project *bitbucket.Project, assignmentResult *AssignmentResult) *ProjectModel {
	return &ProjectModel{
		RetainOnDelete:    plan.RetainOnDelete,
		ID:                types.Int64Value(project.ID),
		Project:           types.StringValue(project.Key),
		Name:              types.StringValue(project.Name),
		Description:       util.NullString(project.Description),
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
