package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type RepositoryModel struct {
	ID              types.String `tfsdk:"id"`
	RetainOnDelete  types.Bool   `tfsdk:"retain_on_delete"`
	ArchiveOnDelete types.Bool   `tfsdk:"archive_on_delete"`
	Project         types.String `tfsdk:"project"`
	Slug            types.String `tfsdk:"slug"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Readme          types.String `tfsdk:"readme"`
	Path            types.String `tfsdk:"path"`

	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ RepositoryPermissionInterface = &RepositoryModel{}

func (m RepositoryModel) getProjectKeyAndSlug(ctx context.Context) (projectKey string, slug string) {
	return m.Project.ValueString(), m.Slug.ValueString()
}

func (m RepositoryModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := m.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func NewRepositoryModel(repository *bitbucket.Repository, plan RepositoryModel, assignmentResult *AssignmentResult) *RepositoryModel {
	return &RepositoryModel{
		ID:                types.StringValue(fmt.Sprintf("%v", repository.ID)),
		Slug:              types.StringValue(repository.Slug),
		Name:              types.StringValue(repository.Name),
		Description:       plan.Description,
		Project:           plan.Project,
		RetainOnDelete:    types.BoolValue(plan.RetainOnDelete.ValueBool()),
		ArchiveOnDelete:   plan.ArchiveOnDelete,
		Readme:            plan.Readme,
		Path:              plan.Path,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
