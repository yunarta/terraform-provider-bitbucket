package provider

import (
	"context"
	"fmt"
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
	_ resource.Resource              = &RepositoryDefaultReviewersResource{}
	_ resource.ResourceWithConfigure = &RepositoryDefaultReviewersResource{}
	_ ConfigurableReceiver           = &RepositoryDefaultReviewersResource{}
)

func NewRepositoryDefaultReviewersResource() resource.Resource {
	return &RepositoryDefaultReviewersResource{}
}

type RepositoryDefaultReviewersResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryDefaultReviewersResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *RepositoryDefaultReviewersResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryDefaultReviewersResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository_default_reviewers"
}

func (receiver *RepositoryDefaultReviewersResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"repository": schema.StringAttribute{
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

func (receiver *RepositoryDefaultReviewersResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *RepositoryDefaultReviewersResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan RepositoryDefaultReviewersModel
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
				Id: refTypesMap[plan.SourceType.ValueString()],
			},
		},
		Reviewers:         reviewers,
		RequiredApprovals: plan.Requires.ValueInt64(),
	}

	reply, err := receiver.client.RepositoryService().AddDefaultReviewers(plan.Project.ValueString(), plan.Repository.ValueString(), defaultReviewers)
	if util.TestError(&response.Diagnostics, err, "Failed to add default reviewer") {
		return
	}

	plan.Id = types.Int64Value(reply.Id)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryDefaultReviewersResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryDefaultReviewersModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	reviewers, err := receiver.client.RepositoryService().ReadDefaultReviewers(state.Project.ValueString(), state.Repository.ValueString(), state.Id.ValueInt64())
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

	state.Source = types.StringValue(reviewers.SourceMatcher.Id)
	state.SourceType = types.StringValue(refTypesReverseMap[reviewers.SourceMatcher.Type.Id])
	state.Target = types.StringValue(reviewers.TargetMatcher.Id)
	state.TargetType = types.StringValue(refTypesReverseMap[reviewers.TargetMatcher.Type.Id])
	state.Reviewers = users
	state.Requires = types.Int64Value(reviewers.RequiredApprovals)

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryDefaultReviewersResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		err         error
		plan, state RepositoryDefaultReviewersModel
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

	err = receiver.client.RepositoryService().UpdateDefaultReviewers(plan.Project.ValueString(), plan.Repository.ValueString(), state.Id.ValueInt64(), defaultReviewers)
	if util.TestError(&response.Diagnostics, err, "Failed to update default reviewer") {
		return
	}

	plan.Id = state.Id

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryDefaultReviewersResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryDefaultReviewersModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err = receiver.client.RepositoryService().DeleteDefaultReviewers(state.Project.ValueString(), state.Repository.ValueString(), state.Id.ValueInt64())
	if util.TestError(&response.Diagnostics, err, "Failed to delete default reviewer") {
		return
	}

	response.State.RemoveResource(ctx)
}
