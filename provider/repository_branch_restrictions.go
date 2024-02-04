package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type RepositoryBranchRestrictions struct {
	ID     types.Int64 `tfsdk:"id"`
	Type   string      `tfsdk:"type"`
	Users  []string    `tfsdk:"users"`
	Groups []string    `tfsdk:"groups"`
}

type RepositoryBranchRestrictionsModel struct {
	Project      string                      `tfsdk:"project"`
	Repository   string                      `tfsdk:"repo"`
	Branch       string                      `tfsdk:"branch"`
	Restrictions []ProjectBranchRestrictions `tfsdk:"restriction"`
}
