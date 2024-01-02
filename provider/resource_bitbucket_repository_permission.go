package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-api-transport/transport"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
	"strings"
)

type RepositoryPermissionsResource struct {
	client *bitbucket.Client
	model  *BitbucketProviderConfig
}

var (
	_ resource.Resource                = &RepositoryPermissionsResource{}
	_ resource.ResourceWithConfigure   = &RepositoryPermissionsResource{}
	_ resource.ResourceWithImportState = &RepositoryPermissionsResource{}
	_ RepositoryPermissionResource     = &RepositoryResource{}
)

func NewRepositoryPermissionsResource() resource.Resource {
	return &RepositoryPermissionsResource{}
}

func (receiver *RepositoryPermissionsResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *RepositoryPermissionsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository_permissions"
}

func (receiver *RepositoryPermissionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"retain_permissions_on_delete": schema.BoolAttribute{
				Optional: true,
			},
			"project": schema.StringAttribute{
				Required: true,
			},
			"slug": schema.StringAttribute{
				Required: true,
			},
			"assignment_version": schema.StringAttribute{
				Optional: true,
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema("REPO_ADMIN", "REPO_READ", "REPO_WRITE"),
		},
	}
}

func (receiver *RepositoryPermissionsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	config, ok := request.ProviderData.(*BitbucketProviderConfig)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AtlassianCloudProviderModel, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	receiver.model = config
	receiver.client = bitbucket.NewBitbucketClient(
		transport.NewHttpPayloadTransport(config.Bitbucket.EndPoint.ValueString(),
			transport.BearerAuthentication{
				Token: config.Bitbucket.Token.ValueString(),
			},
		),
	)
}

func (receiver *RepositoryPermissionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan RepositoryPermissionsModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	computation, diags := CreateRepositoryAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryPermissionsModel(plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryPermissionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state RepositoryPermissionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	computation, diags := ComputeRepositoryAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryPermissionsModel(state, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryPermissionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics

		plan, state RepositoryPermissionsModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateRepositoryAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryPermissionsModel(plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryPermissionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state RepositoryPermissionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainPermissionOnDelete.ValueBool() {
		diags = DeleteRepositoryAssignments(ctx, receiver, state)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *RepositoryPermissionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	slug := strings.Split(request.ID, "/")
	diags := response.State.Set(ctx, &RepositoryModel{
		Project:        types.StringValue(slug[0]),
		Slug:           types.StringValue(slug[1]),
		Assignments:    types.ListNull(assignmentType),
		ComputedUsers:  types.ListNull(computedAssignmentType),
		ComputedGroups: types.ListNull(computedAssignmentType),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
