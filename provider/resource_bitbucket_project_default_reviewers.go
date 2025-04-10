package provider

import (
	"context"
	"fmt"
	"sort"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource              = &ProjectDefaultReviewersResource{}
	_ resource.ResourceWithConfigure = &ProjectDefaultReviewersResource{}
	_ ConfigurableReceiver           = &ProjectDefaultReviewersResource{}
)

func NewProjectDefaultReviewersResource() resource.Resource {
	return &ProjectDefaultReviewersResource{}
}

type ProjectDefaultReviewersResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *ProjectDefaultReviewersResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *ProjectDefaultReviewersResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectDefaultReviewersResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_default_reviewers"
}

func (receiver *ProjectDefaultReviewersResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
			},
			"source": schema.StringAttribute{
				Optional: true,
			},
			"source_type": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(refTypes...),
				},
			},
			"target": schema.StringAttribute{
				Optional: true,
			},
			"target_type": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(refTypes...),
				},
			},
			"reviewers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"requires": schema.Int64Attribute{
				Optional: true,
			},
		},
	}
}

func (receiver *ProjectDefaultReviewersResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectDefaultReviewersResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan ProjectDefaultReviewersModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	reviewers := make([]bitbucket.User, 0)
	for _, user := range plan.Reviewers {
		findUser, _ := receiver.client.UserService().FindUser(user)
		if findUser != nil {
			reviewers = append(reviewers, bitbucket.User{
				Id: findUser.Id,
			})
		}
	}

	defaultReviewers := bitbucket.DefaultReviewers{
		SourceMatcher: bitbucket.SourceMatcher{
			Id: plan.Source.ValueString(),
			Type: bitbucket.DefaultReviewerId{
				Id: refTypesMap[plan.SourceType.ValueString()],
			},
		},
		TargetMatcher: bitbucket.TargetMatcher{
			Id: plan.Target.ValueString(),
			Type: bitbucket.DefaultReviewerId{
				Id: refTypesMap[plan.TargetType.ValueString()],
			},
		},
		Reviewers:         reviewers,
		RequiredApprovals: plan.Requires.ValueInt64(),
	}

	reply, err := receiver.client.ProjectService().AddDefaultReviewers(plan.Project.ValueString(), defaultReviewers)
	if util.TestError(&response.Diagnostics, err, "Failed to add default reviewer") {
		return
	}

	plan.Id = types.Int64Value(reply.Id)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *ProjectDefaultReviewersResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state ProjectDefaultReviewersModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	reviewers, err := receiver.client.ProjectService().ReadDefaultReviewers(state.Project.ValueString(), state.Id.ValueInt64())
	if util.TestError(&response.Diagnostics, err, "Failed get default reviewers") {
		return
	}

	if reviewers == nil {
		(&response.Diagnostics).AddError(
			fmt.Sprintf("Unable to find reviewer for specified %d", state.Id.ValueInt64()),
			"Inconsistence state problem",
		)
		return
	}

	users := make([]string, 0)
	for _, user := range reviewers.Reviewers {
		users = append(users, user.Name)
	}

	sort.Strings(users)
	
	state.Source = types.StringValue(reviewers.SourceMatcher.Id)
	state.SourceType = types.StringValue(refTypesReverseMap[reviewers.SourceMatcher.Type.Id])
	state.Target = types.StringValue(reviewers.TargetMatcher.Id)
	state.TargetType = types.StringValue(refTypesReverseMap[reviewers.TargetMatcher.Type.Id])
	state.Reviewers = users
	state.Requires = types.Int64Value(reviewers.RequiredApprovals)

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (receiver *ProjectDefaultReviewersResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		err         error
		plan, state ProjectDefaultReviewersModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	reviewers := make([]bitbucket.User, 0)
	for _, user := range plan.Reviewers {
		findUser, _ := receiver.client.UserService().FindUser(user)
		if findUser != nil {
			reviewers = append(reviewers, bitbucket.User{
				Id: findUser.Id,
			})
		}
	}

	defaultReviewers := bitbucket.DefaultReviewers{
		Id: state.Id.ValueInt64(),
		SourceMatcher: bitbucket.SourceMatcher{
			Id: plan.Source.ValueString(),
			Type: bitbucket.DefaultReviewerId{
				Id: refTypesMap[plan.SourceType.ValueString()],
			},
		},
		TargetMatcher: bitbucket.TargetMatcher{
			Id: plan.Target.ValueString(),
			Type: bitbucket.DefaultReviewerId{
				Id: refTypesMap[plan.TargetType.ValueString()],
			},
		},
		Reviewers:         reviewers,
		RequiredApprovals: plan.Requires.ValueInt64(),
	}

	err = receiver.client.ProjectService().UpdateDefaultReviewers(plan.Project.ValueString(), state.Id.ValueInt64(), defaultReviewers)
	if util.TestError(&response.Diagnostics, err, "Failed to update default reviewer") {
		return
	}

	plan.Id = state.Id

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *ProjectDefaultReviewersResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state ProjectDefaultReviewersModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err = receiver.client.ProjectService().DeleteDefaultReviewers(state.Project.ValueString(), state.Id.ValueInt64())
	if util.TestError(&response.Diagnostics, err, "Failed to delete default reviewer") {
		return
	}

	response.State.RemoveResource(ctx)
}
