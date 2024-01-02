package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type ProjectPermissionResource interface {
	getClient() *bitbucket.Client
}

type ProjectPermissionInterface interface {
	getAssignment(ctx context.Context) (Assignments, diag.Diagnostics)
	getProjectKey(ctx context.Context) string
}

func CreateProjectAssignments(ctx context.Context, receiver ProjectPermissionResource, plan ProjectPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey := plan.getProjectKey(ctx)
	return ApplyNewAssignmentSet(ctx, receiver.getClient().UserService(),
		*assignmentOrder,
		func(user, requestedPermission string) error {
			return receiver.getClient().ProjectService().UpdateUserPermission(projectKey, user, requestedPermission)
		},
		func(group, requestedPermission string) error {
			return receiver.getClient().ProjectService().UpdateGroupPermission(projectKey, group, requestedPermission)
		},
	)
}

func ComputeProjectAssignments(ctx context.Context, receiver ProjectPermissionResource, state ProjectPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey := state.getProjectKey(ctx)
	assignedPermissions, err := receiver.getClient().ProjectService().ReadPermissions(projectKey)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToReadProjectPermission, err.Error())}
	}

	return ComputeAssignment(ctx, assignedPermissions, *assignmentOrder)
}

func UpdateProjectAssignments(ctx context.Context, receiver ProjectPermissionResource,
	plan ProjectPermissionInterface,
	state ProjectPermissionInterface,
	forceUpdate bool) (*AssignmentResult, diag.Diagnostics) {

	plannedAssignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	inStateAssignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	plannedAssignmentOrder, diags := plannedAssignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	inStateAssignmentOrder, diags := inStateAssignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey := plan.getProjectKey(ctx)

	return UpdateAssignment(ctx, receiver.getClient().UserService(),
		*inStateAssignmentOrder,
		*plannedAssignmentOrder,
		forceUpdate,
		func(user, requestedPermission string) error {
			return receiver.getClient().ProjectService().UpdateUserPermission(projectKey, user, requestedPermission)
		},
		func(group, requestedPermission string) error {
			return receiver.getClient().ProjectService().UpdateGroupPermission(projectKey, group, requestedPermission)
		},
	)
}

func DeleteProjectAssignments(ctx context.Context, receiver ProjectPermissionResource, state ProjectPermissionInterface) diag.Diagnostics {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return diags
	}

	projectKey := state.getProjectKey(ctx)

	assignedPermissions, err := receiver.getClient().ProjectService().ReadPermissions(projectKey)
	if err != nil {
		return []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToReadProjectPermission, err.Error())}
	}

	return RemoveAssignment(ctx, assignedPermissions, *assignmentOrder,
		func(user, requestedPermission string) error {
			return receiver.getClient().ProjectService().UpdateUserPermission(projectKey, user, requestedPermission)
		},
		func(group, requestedPermission string) error {
			return receiver.getClient().ProjectService().UpdateGroupPermission(projectKey, group, requestedPermission)
		})
}
