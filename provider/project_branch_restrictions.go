package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProjectBranchRestrictions struct {
	ID     types.Int64 `tfsdk:"id"`
	Type   string      `tfsdk:"type"`
	Users  []string    `tfsdk:"users"`
	Groups []string    `tfsdk:"groups"`
}

type ProjectBranchRestrictionsModel struct {
	Project      string                      `tfsdk:"project"`
	Branch       string                      `tfsdk:"branch"`
	Restrictions []ProjectBranchRestrictions `tfsdk:"restriction"`
}
