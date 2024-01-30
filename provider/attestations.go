package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
	"sort"
)

func CreateAttestation(ctx context.Context, permissions *bitbucket.ObjectPermission, diagnostics *diag.Diagnostics) (basetypes.MapValue, basetypes.MapValue, diag.Diagnostics) {
	var userPermissionsMap = make(map[string][]string)
	var groupPermissionsMap = make(map[string][]string)
	for _, user := range permissions.Users {
		userInPermission, ok := userPermissionsMap[user.Permission]
		if !ok {
			userInPermission = make([]string, 0)
			userPermissionsMap[user.Permission] = userInPermission
		}

		userInPermission = append(userInPermission, user.Owner.Name)
		userPermissionsMap[user.Permission] = userInPermission
	}

	for _, group := range permissions.Groups {
		groupInPermission, ok := groupPermissionsMap[group.Permission]
		if !ok {
			groupInPermission = make([]string, 0)
		}

		groupInPermission = append(groupInPermission, group.Owner.Name)
		groupPermissionsMap[group.Permission] = groupInPermission
	}

	for _, groups := range groupPermissionsMap {
		sort.Strings(groups)
	}

	for _, users := range userPermissionsMap {
		sort.Strings(users)
	}

	users, diags := types.MapValueFrom(ctx, types.ListType{
		ElemType: types.StringType,
	}, userPermissionsMap)
	if util.TestDiagnostic(diagnostics, diags) {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags
	}

	groups, diags := types.MapValueFrom(ctx, types.ListType{
		ElemType: types.StringType,
	}, groupPermissionsMap)
	if util.TestDiagnostic(diagnostics, diags) {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags
	}

	return users, groups, nil
}
