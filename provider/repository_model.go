package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type RepositoryModel struct {
	ID             types.String `tfsdk:"id"`
	RetainOnDelete bool         `tfsdk:"retain_on_delete"`
	Project        string       `tfsdk:"project"`
	Slug           types.String `tfsdk:"slug"`
	Name           string       `tfsdk:"name"`
	Description    string       `tfsdk:"description"`
	Readme         types.String `tfsdk:"readme"`
	Path           types.String `tfsdk:"path"`

	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       Assignments  `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ RepositoryPermissionInterface = &RepositoryModel{}

func (m RepositoryModel) getProjectKeyAndSlug(ctx context.Context) (projectKey string, slug string) {
	return m.Project, m.Slug.ValueString()
}

func (m RepositoryModel) getAssignment(ctx context.Context) Assignments {
	return m.Assignments
}

func NewRepositoryModel(repository *bitbucket.Repository, plan RepositoryModel, assignmentResult *AssignmentResult) *RepositoryModel {
	return &RepositoryModel{
		ID:                types.StringValue(fmt.Sprintf("%v", repository.ID)),
		Slug:              types.StringValue(repository.Slug),
		Name:              plan.Name,
		Description:       plan.Description,
		Project:           plan.Project,
		RetainOnDelete:    plan.RetainOnDelete,
		Readme:            plan.Readme,
		Path:              plan.Path,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
