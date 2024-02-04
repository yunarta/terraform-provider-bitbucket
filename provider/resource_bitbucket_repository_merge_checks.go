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
	_ resource.Resource              = &RepositoryMergeChecksResource{}
	_ resource.ResourceWithConfigure = &RepositoryMergeChecksResource{}
	_ ConfigurableReceiver           = &RepositoryMergeChecksResource{}
)

func NewRepositoryMergeChecksResource() resource.Resource {
	return &RepositoryMergeChecksResource{}
}

type RepositoryMergeChecksResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryMergeChecksResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *RepositoryMergeChecksResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryMergeChecksResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository_merge_checks"
}

func (receiver *RepositoryMergeChecksResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"repo": schema.StringAttribute{
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

func (receiver *RepositoryMergeChecksResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *RepositoryMergeChecksResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan RepositoryMergeChecksModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if plan.AllReviewerApproval.ValueBool() {
		err = receiver.client.RepositoryService().EnableMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
		if util.TestError(&response.Diagnostics, err, "Failed to update all approvers merge check") {
			return
		}
	}

	if !plan.MinimumApproval.IsNull() {
		err = receiver.client.RepositoryService().ConfigureMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook", int(plan.MinimumApproval.ValueInt64()))
		if util.TestError(&response.Diagnostics, err, "Failed to update minimum approvers merge check") {
			return
		}
	}

	if !plan.MinimumSuccessfulBuild.IsNull() {
		err = receiver.client.RepositoryService().ConfigureMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck", int(plan.MinimumSuccessfulBuild.ValueInt64()))
		if util.TestError(&response.Diagnostics, err, "Failed to update minimum build merge check") {
			return
		}
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryMergeChecksResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryMergeChecksModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	checks, err := receiver.client.RepositoryService().GetMergeChecks(state.Project, state.Repo)
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
		settings, err := receiver.client.RepositoryService().GetMergeCheckSetting(state.Project, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook")
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
		settings, err := receiver.client.RepositoryService().GetMergeCheckSetting(state.Project, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck")
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

func (receiver *RepositoryMergeChecksResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		err         error
		plan, state RepositoryMergeChecksModel
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
		err = receiver.client.RepositoryService().EnableMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	} else {
		err = receiver.client.RepositoryService().DisableMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	}
	if util.TestError(&response.Diagnostics, err, "Failed to update all approvers merge check") {
		return
	}

	if !plan.MinimumApproval.IsNull() {
		err = receiver.client.RepositoryService().ConfigureMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook", int(plan.MinimumApproval.ValueInt64()))
	} else {
		err = receiver.client.RepositoryService().DisableMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:requiredApproversMergeHook")
	}
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum approvers merge check") {
		return
	}

	if !plan.MinimumSuccessfulBuild.IsNull() {
		err = receiver.client.RepositoryService().ConfigureMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck", int(plan.MinimumSuccessfulBuild.ValueInt64()))
	} else {
		err = receiver.client.RepositoryService().DisableMergeCheck(plan.Project, plan.Repo, "com.atlassian.bitbucket.server.bitbucket-build:requiredBuildsMergeCheck")
	}
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum build merge check") {
		return
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryMergeChecksResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryMergeChecksModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err = receiver.client.RepositoryService().DisableMergeCheck(state.Project, state.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	if util.TestError(&response.Diagnostics, err, "Failed to update all approvers merge check") {
		return
	}

	err = receiver.client.RepositoryService().DisableMergeCheck(state.Project, state.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum approvers merge check") {
		return
	}

	err = receiver.client.RepositoryService().DisableMergeCheck(state.Project, state.Repo, "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:all-approvers-merge-check")
	if util.TestError(&response.Diagnostics, err, "Failed to update minimum build merge check") {
		return
	}

	response.State.RemoveResource(ctx)
}
