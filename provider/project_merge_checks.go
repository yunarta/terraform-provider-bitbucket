package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProjectMergeChecksModel struct {
	Project                string      `tfsdk:"project"`
	AllReviewerApproval    types.Bool  `tfsdk:"all_reviewer_approval"`
	MinimumApproval        types.Int64 `tfsdk:"minimum_approvals"`
	MinimumSuccessfulBuild types.Int64 `tfsdk:"minimum_successful_builds"`
}
