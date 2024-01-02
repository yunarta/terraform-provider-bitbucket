package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
)

type RepositoryModel struct {
	ID                       types.String `tfsdk:"id"`
	RetainOnDelete           types.Bool   `tfsdk:"retain_on_delete"`
	RetainPermissionOnDelete types.Bool   `tfsdk:"retain_permissions_on_delete"`
	Project                  types.String `tfsdk:"project"`
	Slug                     types.String `tfsdk:"slug"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Readme                   types.String `tfsdk:"readme"`

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
		ID:                       types.StringValue(fmt.Sprintf("%v", repository.ID)),
		Slug:                     types.StringValue(repository.Slug),
		Name:                     types.StringValue(repository.Name),
		Description:              util.NullString(repository.Description),
		Project:                  types.StringValue(repository.Project.Key),
		RetainOnDelete:           plan.RetainOnDelete,
		RetainPermissionOnDelete: plan.RetainPermissionOnDelete,
		Readme:                   plan.Readme,
		AssignmentVersion:        plan.AssignmentVersion,
		Assignments:              plan.Assignments,
		ComputedUsers:            assignmentResult.ComputedUsers,
		ComputedGroups:           assignmentResult.ComputedGroups,
	}
}
