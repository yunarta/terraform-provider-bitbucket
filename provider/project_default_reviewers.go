package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProjectDefaultReviewersModel struct {
	Id         types.Int64  `tfsdk:"id"`
	Project    types.String `tfsdk:"project"`
	Source     types.String `tfsdk:"source"`
	SourceType types.String `tfsdk:"source_type"`
	Target     types.String `tfsdk:"target"`
	TargetType types.String `tfsdk:"target_type"`
	Reviewers  []string     `tfsdk:"reviewers"`
	Requires   types.Int64  `tfsdk:"requires"`
}
