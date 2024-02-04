package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
	"strconv"
)

var (
	_ resource.Resource              = &ProjectMergeChecksResource{}
	_ resource.ResourceWithConfigure = &ProjectMergeChecksResource{}
	_ ConfigurableReceiver           = &ProjectMergeChecksResource{}
)

func NewProjectMergeChecksResource() resource.Resource {
	return &ProjectMergeChecksResource{}
}

type ProjectMergeChecksResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *ProjectMergeChecksResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *ProjectMergeChecksResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectMergeChecksResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_merge_checks"
}

func (receiver *ProjectMergeChecksResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"all_reviewer_approval": schema.BoolAttribute{
				Optional: true,
			},
			"minimum_approvals": schema.Int64Attribute{
				Optional: true,
			},
			"minimum_successful_builds": schema.Int64Attribute{
				Optional: true,
			},
		},
	}
}

func (receiver *ProjectMergeChecksResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectMergeChecksResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan ProjectMergeChecksModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if plan.AllReviewerApproval.ValueBool() {
		err = receiver.client.ProjectService().EnableMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
		if util.TestError(&response.Diagnostics, err, "Failed to update all approvers merge check") {
			return
		}
	}

	if !plan.MinimumApproval.IsNull() {
		err = receiver.client.ProjectService().ConfigureMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook", int(plan.MinimumApproval.ValueInt64()))
		if util.TestError(&response.Diagnostics, err, "Failed to update minimum approvers merge check") {
			return
		}
	}

	if !plan.MinimumSuccessfulBuild.IsNull() {
		err = receiver.client.ProjectService().ConfigureMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck", int(plan.MinimumSuccessfulBuild.ValueInt64()))
		if util.TestError(&response.Diagnostics, err, "Failed to update minimum build merge check") {
			return
		}
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *ProjectMergeChecksResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state ProjectMergeChecksModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	checks, err := receiver.client.ProjectService().GetMergeChecks(state.Project)
	if util.TestError(&response.Diagnostics, err, "Failed get merge checks") {
		return
	}

	// convert checks into map using it detail key
	checksMap := make(map[string]bitbucket.MergeCheck)
	for _, check := range checks {
		checksMap[check.Details.Key] = check
	}

	allReviewerApproval := checksMap["com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check"]
	state.AllReviewerApproval = types.BoolValue(allReviewerApproval.Enabled)

	minimumApproval := checksMap["com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check"]
	if minimumApproval.Enabled {
		settings, err := receiver.client.ProjectService().GetMergeCheckSetting(state.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook")
		if util.TestError(&response.Diagnostics, err, "Failed get update minimum approvers merge check") {
			return
		}

		atoi, err := strconv.Atoi(settings.RequiredCount)
		state.MinimumApproval = types.Int64Value(int64(atoi))
	} else {
		state.MinimumApproval = types.Int64Null()
	}

	minimumBuild := checksMap["com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck"]
	if minimumBuild.Enabled {
		settings, err := receiver.client.ProjectService().GetMergeCheckSetting(state.Project, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck")
		if util.TestError(&response.Diagnostics, err, "Failed to get minimum approvers merge check") {
			return
		}

		atoi, err := strconv.Atoi(settings.RequiredCount)
		state.MinimumSuccessfulBuild = types.Int64Value(int64(atoi))
	} else {
		state.MinimumSuccessfulBuild = types.Int64Null()
	}

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (receiver *ProjectMergeChecksResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		err         error
		plan, state ProjectMergeChecksModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if plan.AllReviewerApproval.ValueBool() {
		err = receiver.client.ProjectService().EnableMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	} else {
		err = receiver.client.ProjectService().DisableMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	}
	if util.TestError(&response.Diagnostics, err, "Failed to update all approvers merge check") {
		return
	}

	if !plan.MinimumApproval.IsNull() {
		err = receiver.client.ProjectService().ConfigureMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook", int(plan.MinimumApproval.ValueInt64()))
	} else {
		err = receiver.client.ProjectService().DisableMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook")
	}
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum approvers merge check") {
		return
	}

	if !plan.MinimumSuccessfulBuild.IsNull() {
		err = receiver.client.ProjectService().ConfigureMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck", int(plan.MinimumSuccessfulBuild.ValueInt64()))
	} else {
		err = receiver.client.ProjectService().DisableMergeCheck(plan.Project, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck")
	}
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum build merge check") {
		return
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *ProjectMergeChecksResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state ProjectMergeChecksModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err = receiver.client.ProjectService().DisableMergeCheck(state.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	if util.TestError(&response.Diagnostics, err, "Failed to update all approvers merge check") {
		return
	}

	err = receiver.client.ProjectService().DisableMergeCheck(state.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum approvers merge check") {
		return
	}

	err = receiver.client.ProjectService().DisableMergeCheck(state.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum build merge check") {
		return
	}

	response.State.RemoveResource(ctx)
}
