package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
	"sort"
)

var (
	_ resource.Resource              = &RepositoryBranchRestrictionsResource{}
	_ resource.ResourceWithConfigure = &RepositoryBranchRestrictionsResource{}
	_ ConfigurableReceiver           = &RepositoryBranchRestrictionsResource{}
)

func NewRepositoryBranchRestrictionsResource() resource.Resource {
	return &RepositoryBranchRestrictionsResource{}
}

type RepositoryBranchRestrictionsResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryBranchRestrictionsResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *RepositoryBranchRestrictionsResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryBranchRestrictionsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository_branch_restrictions"
}

func (receiver *RepositoryBranchRestrictionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
			},
			"repo": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
			},
			"branch": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"restriction": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
						"users": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"groups": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (receiver *RepositoryBranchRestrictionsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *RepositoryBranchRestrictionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		plan  RepositoryBranchRestrictionsModel
		err   error
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Read(plan.Project, plan.Repository)
	if util.TestError(&response.Diagnostics, err, "Failed to read project") {
		return
	}

	branchRestrictions := make([]bitbucket.BranchRestriction, 0)
	for _, restriction := range plan.Restrictions {
		branchRestriction := bitbucket.BranchRestriction{
			Matcher: bitbucket.BranchRestrictionMatcher{
				Id:        plan.Branch,
				DisplayId: plan.Branch,
				Type: bitbucket.BranchRestrictionMatcherType{
					Id: "BRANCH",
				},
			},
			Scope: bitbucket.BranchRestrictionScope{
				Type:       "REPOSITORY",
				ResourceId: repository.ID,
			},
			Type:   restriction.Type,
			Users:  restriction.Users,
			Groups: restriction.Groups,
		}
		branchRestrictions = append(branchRestrictions, branchRestriction)
	}

	branchRestrictionsReply, err := receiver.client.RepositoryService().CreateBranchRestrictions(plan.Project, plan.Repository, branchRestrictions)
	if util.TestError(&response.Diagnostics, err, "Failed to create branch restrictions") {
		return
	}

	// iterate branchRestrictions, get the id, put into plan id
	for i, restriction := range branchRestrictionsReply {
		plan.Restrictions[i].ID = types.Int64Value(restriction.Id)
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryBranchRestrictionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state RepositoryBranchRestrictionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for i, restriction := range state.Restrictions {
		branchRestriction, err := receiver.client.RepositoryService().ReadBranchRestriction(state.Project, state.Repository, restriction.ID.ValueInt64())
		if util.TestError(&response.Diagnostics, err, "Failed to read branch restriction") {
			return
		}

		state.Restrictions[i].Type = branchRestriction.Type
		if len(branchRestriction.Users) > 0 {
			state.Restrictions[i].Users = make([]string, len(branchRestriction.Users))
			for j, user := range branchRestriction.Users {
				state.Restrictions[i].Users[j] = user.Name
			}
			sort.Strings(state.Restrictions[i].Users)
		}

		if len(branchRestriction.Groups) > 0 {
			state.Restrictions[i].Groups = branchRestriction.Groups
		}
	}

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryBranchRestrictionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		plan, state RepositoryBranchRestrictionsModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Read(plan.Project, plan.Repository)
	if util.TestError(&response.Diagnostics, err, "Failed to read project") {
		return
	}

	// collect the state resctrion id, we wanted to delete it
	// collect the state resctrion id, we wanted to delete it
	stateRestrictionIDs := make(map[int64]bool)
	for _, restriction := range state.Restrictions {
		stateRestrictionIDs[restriction.ID.ValueInt64()] = false
	}

	branchRestrictions := make([]bitbucket.BranchRestriction, 0)
	for _, restriction := range plan.Restrictions {
		branchRestriction := bitbucket.BranchRestriction{
			Id: int(restriction.ID.ValueInt64()),
			Matcher: bitbucket.BranchRestrictionMatcher{
				Id:        plan.Branch,
				DisplayId: plan.Branch,
				Type: bitbucket.BranchRestrictionMatcherType{
					Id:   "BRANCH",
					Name: "branch",
				},
			},
			Scope: bitbucket.BranchRestrictionScope{
				Type:       "REPOSITORY",
				ResourceId: repository.ID,
			},
			Type:   restriction.Type,
			Users:  restriction.Users,
			Groups: restriction.Groups,
		}
		branchRestrictions = append(branchRestrictions, branchRestriction)
	}

	branchRestrictionsReply, err := receiver.client.RepositoryService().CreateBranchRestrictions(plan.Project, plan.Repository, branchRestrictions)
	if util.TestError(&response.Diagnostics, err, "Failed to create branch restrictions") {
		return
	}

	// iterate the reply, if the id is not i stateRestrictionIDs[restriction.ID.ValueInt64()] = true, then delete
	// iterate the reply, if the id is not in stateRestrictionIDs, then delete
	for i, restriction := range branchRestrictionsReply {
		plan.Restrictions[i].ID = types.Int64Value(restriction.Id)
		stateRestrictionIDs[restriction.Id] = true
	}
	// delete the branch restrictions that are not in the state
	for id, keep := range stateRestrictionIDs {
		if !keep {
			err = receiver.client.RepositoryService().DeleteBranchRestriction(plan.Project, plan.Repository, id)
			if util.TestError(&response.Diagnostics, err, "Failed to delete branch restriction") {
				return
			}
		}
	}

	// iterate branchRestrictions, get the id, put into plan id
	for i, restriction := range branchRestrictionsReply {
		plan.Restrictions[i].ID = types.Int64Value(restriction.Id)
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (receiver *RepositoryBranchRestrictionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics

		state RepositoryBranchRestrictionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, restriction := range state.Restrictions {
		err := receiver.client.RepositoryService().DeleteBranchRestriction(state.Project, state.Repository, restriction.ID.ValueInt64())
		if util.TestError(&response.Diagnostics, err, "Failed to read branch restriction") {
			return
		}
	}

	response.State.RemoveResource(ctx)
}
