package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type RepositoryDefaultReviewersModel struct {
	Id         types.Int64  `tfsdk:"id"`
	Project    types.String `tfsdk:"project"`
	Repository types.String `tfsdk:"repository"`
	Source     types.String `tfsdk:"source"`
	SourceType types.String `tfsdk:"source_type"`
	Target     types.String `tfsdk:"target"`
	TargetType types.String `tfsdk:"target_type"`
	Reviewers  []string     `tfsdk:"reviewers"`
	Requires   types.Int64  `tfsdk:"requires"`
}
