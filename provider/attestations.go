package provider

import (
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"sort"
)

func CreateAttestation(permissions *bitbucket.ObjectPermission, availablePermissions []string) (map[string][]string, map[string][]string) {
	var userPermissionsMap = make(map[string][]string)
	var groupPermissionsMap = make(map[string][]string)
	for _, permission := range availablePermissions {
		userPermissionsMap[permission] = []string{}
		groupPermissionsMap[permission] = []string{}
	}

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

	return userPermissionsMap, userPermissionsMap
}
