package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource                = &ProjectPermissionsResource{}
	_ resource.ResourceWithConfigure   = &ProjectPermissionsResource{}
	_ resource.ResourceWithImportState = &ProjectPermissionsResource{}
	_ ProjectPermissionResource        = &ProjectPermissionsResource{}
	_ ConfigurableReceiver             = &ProjectPermissionsResource{}
)

func NewProjectPermissionsResource() resource.Resource {
	return &ProjectPermissionsResource{}
}

type ProjectPermissionsResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *ProjectPermissionsResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *ProjectPermissionsResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectPermissionsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_permissions"
}

func (receiver *ProjectPermissionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"assignment_version": schema.StringAttribute{
				Optional: true,
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema("PROJECT_ADMIN", "REPO_CREATE", "PROJECT_READ", "PROJECT_WRITE"),
		},
	}
}

func (receiver *ProjectPermissionsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectPermissionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		plan  ProjectPermissionsModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	computation, diags := CreateProjectAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectPermissionsModel(plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectPermissionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state ProjectPermissionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	computation, diags := ComputeProjectAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectPermissionsModel(state, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectPermissionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		plan, state ProjectPermissionsModel
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
	computation, diags := UpdateProjectAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectPermissionsModel(plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectPermissionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state ProjectPermissionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		diags = DeleteProjectAssignments(ctx, receiver, state)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *ProjectPermissionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	diags := response.State.Set(ctx, &ProjectPermissionsModel{
		Key:            types.StringValue(request.ID),
		Assignments:    types.ListNull(assignmentType),
		ComputedUsers:  types.ListNull(computedAssignmentType),
		ComputedGroups: types.ListNull(computedAssignmentType),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
