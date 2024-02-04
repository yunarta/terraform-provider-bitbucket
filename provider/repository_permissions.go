package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type RepositoryPermissionReceiver interface {
	getClient() *bitbucket.Client
}

type RepositoryPermissionInterface interface {
	getAssignment(ctx context.Context) Assignments
	getProjectKeyAndSlug(ctx context.Context) (projectKey string, slug string)
}

func CreateRepositoryAssignments(ctx context.Context, receiver RepositoryPermissionReceiver, plan RepositoryPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments := plan.getAssignment(ctx)

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey, slug := plan.getProjectKeyAndSlug(ctx)
	return ApplyNewAssignmentSet(ctx, receiver.getClient().UserService(),
		*assignmentOrder,
		func(user, requestedPermission string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermission(projectKey, slug, user, requestedPermission)
		},
		func(group, requestedPermission string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermission(projectKey, slug, group, requestedPermission)
		},
	)
}

func ComputeRepositoryAssignments(ctx context.Context, receiver RepositoryPermissionReceiver, state RepositoryPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments := state.getAssignment(ctx)

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey, slug := state.getProjectKeyAndSlug(ctx)
	assignedPermissions, err := receiver.getClient().RepositoryService().ReadPermissions(projectKey, slug)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToReadRepositoryPermission, err.Error())}
	}

	return ComputeAssignment(ctx, assignedPermissions, *assignmentOrder)
}

func UpdateRepositoryAssignments(ctx context.Context, receiver RepositoryPermissionReceiver,
	plan RepositoryPermissionInterface,
	state RepositoryPermissionInterface,
	forceUpdate bool) (*AssignmentResult, diag.Diagnostics) {

	plannedAssignments := plan.getAssignment(ctx)

	inStateAssignments := state.getAssignment(ctx)

	plannedAssignmentOrder, diags := plannedAssignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	inStateAssignmentOrder, diags := inStateAssignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey, slug := state.getProjectKeyAndSlug(ctx)

	return UpdateAssignment(ctx, receiver.getClient().UserService(),
		*inStateAssignmentOrder,
		*plannedAssignmentOrder,
		forceUpdate,
		func(user, requestedPermission string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermission(projectKey, slug, user, requestedPermission)
		},
		func(group, requestedPermission string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermission(projectKey, slug, group, requestedPermission)
		},
	)
}

func DeleteRepositoryAssignments(ctx context.Context, receiver RepositoryPermissionReceiver, state RepositoryPermissionInterface) diag.Diagnostics {
	assignments := state.getAssignment(ctx)

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return diags
	}

	projectKey, slug := state.getProjectKeyAndSlug(ctx)

	assignedPermissions, err := receiver.getClient().RepositoryService().ReadPermissions(projectKey, slug)
	if err != nil {
		return []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToReadRepositoryPermission, err.Error())}
	}

	return RemoveAssignment(ctx, assignedPermissions, *assignmentOrder,
		func(user, requestedPermission string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermission(projectKey, slug, user, requestedPermission)
		},
		func(group, requestedPermission string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermission(projectKey, slug, group, requestedPermission)
		})
}
