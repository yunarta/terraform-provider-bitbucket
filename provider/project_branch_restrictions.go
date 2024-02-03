package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProjectBranchRestriction struct {
	ID     types.Int64 `tfsdk:"id"`
	Type   string      `tfsdk:"type"`
	Users  []string    `tfsdk:"users"`
	Groups []string    `tfsdk:"groups"`
}

type ProjectBranchRestrictionModel struct {
	Project      string                     `tfsdk:"project"`
	Branch       string                     `tfsdk:"branch"`
	Restrictions []ProjectBranchRestriction `tfsdk:"restriction"`
}
