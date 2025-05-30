package provider

import (
	"context"
	"slices"
	"strings"

	"github.com/emirpasic/gods/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yunarta/golang-quality-of-life-pack/collections"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type Assignment struct {
	Users      []string `tfsdk:"users"`
	Groups     []string `tfsdk:"groups"`
	Permission string   `tfsdk:"permission"`
	Priority   int64    `tfsdk:"priority"`
}

type AssignmentOrder struct {
	Users      map[string]string
	UserNames  []string
	Groups     map[string]string
	GroupNames []string
}

type Assignments []Assignment

type UpdateUserPermissionFunc func(user string, requestedPermission string) error
type UpdateGroupPermissionFunc func(group string, requestedPermission string) error

func (assignments Assignments) CreateAssignmentOrder(ctx context.Context) (*AssignmentOrder, diag.Diagnostics) {
	var priorities []int64
	var makeAssignments = map[int64]Assignment{}
	for _, assignment := range assignments {
		priorities = append(priorities, assignment.Priority)
		makeAssignments[assignment.Priority] = assignment
	}
	slices.SortFunc(priorities, func(a, b int64) int {
		return -utils.Int64Comparator(a, b)
	})

	var usersAssignments = map[string]string{}
	var groupsAssignments = map[string]string{}
	var userNames = make([]string, 0)
	var groupNames = make([]string, 0)
	for _, priority := range priorities {
		assignment := makeAssignments[priority]
		for _, user := range assignment.Users {
			usersAssignments[user] = assignment.Permission
			userNames = append(userNames, user)
		}

		for _, group := range assignment.Groups {
			groupsAssignments[group] = assignment.Permission
			groupNames = append(groupNames, group)
		}
	}

	return &AssignmentOrder{
		Users:      usersAssignments,
		UserNames:  userNames,
		Groups:     groupsAssignments,
		GroupNames: groupNames,
	}, nil
}

func AssignmentSchema(permissions ...string) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"users": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"groups": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"permission": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf(permissions...),
					},
				},
				"priority": schema.Int64Attribute{
					Required: true,
				},
			},
		},
	}
}

var ComputedAssignmentSchema = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed: true,
			},
			"permission": schema.StringAttribute{
				Computed: true,
			},
		},
	},
}

type ComputedAssignment struct {
	Name       string `tfsdk:"name"`
	Permission string `tfsdk:"permission"`
}

var assignmentType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"permission": types.StringType,
		"priority":   types.NumberType,
		"users": types.ListType{
			ElemType: types.StringType,
		},
		"groups": types.ListType{
			ElemType: types.StringType,
		},
	},
}

var computedAssignmentType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"permission": types.StringType,
		"name":       types.StringType,
	},
}

type AssignmentResult struct {
	ComputedUsers  types.List
	ComputedGroups types.List
}

func ApplyNewAssignmentSet(ctx context.Context, userService *bitbucket.UserService,
	assignmentOrder AssignmentOrder,
	updateUserPermission UpdateUserPermissionFunc,
	updateGroupPermission UpdateGroupPermissionFunc) (*AssignmentResult, diag.Diagnostics) {

	computedUsers := make([]ComputedAssignment, 0)
	computedGroups := make([]ComputedAssignment, 0)

	for user, requestedPermission := range assignmentOrder.Users {
		found, err := userService.FindUser(user)
		if found == nil {
			continue
		}

		computedUsers = append(computedUsers, ComputedAssignment{
			Name:       user,
			Permission: requestedPermission,
		})

		err = updateUserPermission(user, requestedPermission)
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToUpdateUserPermission, err.Error())}
		}
	}

	for group, requestedPermission := range assignmentOrder.Groups {
		found, err := userService.FindGroup(group)
		if found == nil {
			continue
		}

		computedGroups = append(computedGroups, ComputedAssignment{
			Name:       group,
			Permission: requestedPermission,
		})

		err = updateGroupPermission(group, requestedPermission)
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToUpdateGroupPermission, err.Error())}
		}
	}

	return createAssignmentResult(ctx, computedUsers, computedGroups)
}

func UpdateAssignment(ctx context.Context, userService *bitbucket.UserService,
	inStateAssignmentOrder AssignmentOrder,
	plannedAssignmentOrder AssignmentOrder,
	forceUpdate bool,
	updateUserPermission UpdateUserPermissionFunc,
	updateGroupPermission UpdateGroupPermissionFunc) (*AssignmentResult, diag.Diagnostics) {

	computedUsers, diags := updateUsers(inStateAssignmentOrder, plannedAssignmentOrder, userService, forceUpdate, updateUserPermission)
	if diags != nil {
		return nil, diags
	}

	computedGroups, diags := updateGroups(inStateAssignmentOrder, plannedAssignmentOrder, userService, forceUpdate, updateGroupPermission)
	if diags != nil {
		return nil, diags
	}

	return createAssignmentResult(ctx, computedUsers, computedGroups)
}

func updateUsers(inStateAssignmentOrder AssignmentOrder, plannedAssignmentOrder AssignmentOrder,
	userService *bitbucket.UserService, forceUpdate bool, updateUserPermission UpdateUserPermissionFunc) ([]ComputedAssignment, diag.Diagnostics) {

	computedUsers := make([]ComputedAssignment, 0)

	_, removing := collections.Delta(inStateAssignmentOrder.UserNames, plannedAssignmentOrder.UserNames)
	for _, user := range plannedAssignmentOrder.UserNames {
		if collections.Contains(removing, user) {
			continue
		}

		found, err := userService.FindUser(user)
		if found == nil {
			continue
		}

		requestedPermission := plannedAssignmentOrder.Users[user]
		inStatePermission := inStateAssignmentOrder.Users[user]
		computedUsers = append(computedUsers, ComputedAssignment{
			Name:       user,
			Permission: requestedPermission,
		})

		if inStatePermission != requestedPermission || forceUpdate {
			err = updateUserPermission(user, requestedPermission)
			if err != nil {
				return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToUpdateUserPermission, err.Error())}
			}
		}
	}

	for _, user := range removing {
		err := updateUserPermission(user, "")
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToRemoveUserPermission, err.Error())}
		}
	}

	return computedUsers, nil
}

func updateGroups(inStateAssignmentOrder AssignmentOrder, plannedAssignmentOrder AssignmentOrder,
	userService *bitbucket.UserService, forceUpdate bool, updateGroupPermission UpdateGroupPermissionFunc) ([]ComputedAssignment, diag.Diagnostics) {

	computedGroups := make([]ComputedAssignment, 0)

	_, removing := collections.Delta(inStateAssignmentOrder.GroupNames, plannedAssignmentOrder.GroupNames)
	for _, group := range plannedAssignmentOrder.GroupNames {
		if collections.Contains(removing, group) {
			continue
		}

		found, err := userService.FindGroup(group)
		if found == nil {
			continue
		}

		requestedPermission := plannedAssignmentOrder.Groups[group]
		inStatePermission := inStateAssignmentOrder.Groups[group]
		computedGroups = append(computedGroups, ComputedAssignment{
			Name:       group,
			Permission: requestedPermission,
		})

		if inStatePermission != requestedPermission || forceUpdate {
			err = updateGroupPermission(found.Name, requestedPermission)
			if err != nil {
				return nil, []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToUpdateGroupPermission, err.Error())}
			}
		}
	}

	for _, group := range removing {
		found, err := userService.FindGroup(group)
		if found == nil {
			continue
		}

		err := updateGroupPermission(found.Name, "")
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to remove group permission", err.Error())}
		}
	}

	return computedGroups, nil
}

func RemoveAssignment(ctx context.Context,
	assignedPermissions *bitbucket.ObjectPermission, assignmentOrder AssignmentOrder,
	updateUserPermission UpdateUserPermissionFunc,
	updateGroupPermission UpdateGroupPermissionFunc) diag.Diagnostics {

	for _, user := range assignedPermissions.Users {
		if _, ok := assignmentOrder.Users[user.Owner.Name]; ok {
			err := updateUserPermission(user.Owner.Name, "")
			if err != nil {
				return []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToRemoveUserPermission, err.Error())}
			}
		}
	}

	for _, group := range assignedPermissions.Groups {
		if _, ok := assignmentOrder.Groups[strings.ToLower(group.Owner.Name)]; ok {
			err := updateGroupPermission(group.Owner.Name, "")
			if err != nil {
				return []diag.Diagnostic{diag.NewErrorDiagnostic(errorFailedToRemoveGroupPermission, err.Error())}
			}
		}
	}

	return nil
}

func ComputeAssignment(ctx context.Context,
	assignedPermissions *bitbucket.ObjectPermission, assignmentOrder AssignmentOrder) (*AssignmentResult, diag.Diagnostics) {

	computedUsers := make([]ComputedAssignment, 0)
	computedGroups := make([]ComputedAssignment, 0)

	for _, user := range assignedPermissions.Users {
		if _, ok := assignmentOrder.Users[user.Owner.Name]; ok {
			computedUsers = append(computedUsers, ComputedAssignment{
				Name:       user.Owner.Name,
				Permission: user.Permission,
			})
		}
	}

	for _, group := range assignedPermissions.Groups {
		if _, ok := assignmentOrder.Groups[strings.ToLower(group.Owner.Name)]; ok {
			computedGroups = append(computedGroups, ComputedAssignment{
				Name:       group.Owner.Name,
				Permission: group.Permission,
			})
		}
	}

	return createAssignmentResult(ctx, computedUsers, computedGroups)
}

func createAssignmentResult(ctx context.Context, computedUsers []ComputedAssignment, computedGroups []ComputedAssignment) (*AssignmentResult, diag.Diagnostics) {
	computedUsersList, diags := createTfList(ctx, computedUsers)
	if diags != nil {
		return nil, diags
	}

	computedGroupsList, diags := createTfList(ctx, computedGroups)
	if diags != nil {
		return nil, diags
	}

	return &AssignmentResult{
		ComputedUsers:  *computedUsersList,
		ComputedGroups: *computedGroupsList,
	}, nil
}

func createTfList(ctx context.Context, assignments []ComputedAssignment) (*basetypes.ListValue, diag.Diagnostics) {
	slices.SortFunc(assignments, func(a, b ComputedAssignment) int {
		return strings.Compare(a.Name, b.Name)
	})

	computedUsersList, diags := types.ListValueFrom(ctx, computedAssignmentType, assignments)
	if diags != nil {
		return nil, diags
	}

	return &computedUsersList, nil
}
