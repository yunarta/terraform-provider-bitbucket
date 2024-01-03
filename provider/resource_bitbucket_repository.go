package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
	"regexp"
	"strconv"
	"strings"
)

var (
	_ resource.Resource                = &RepositoryResource{}
	_ resource.ResourceWithConfigure   = &RepositoryResource{}
	_ resource.ResourceWithImportState = &RepositoryResource{}
	_ RepositoryPermissionResource     = &RepositoryResource{}
	_ ConfigurableReceiver             = &RepositoryResource{}
)

func NewRepositoryResource() resource.Resource {
	return &RepositoryResource{}
}

type RepositoryResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *RepositoryResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository"
}

func (receiver *RepositoryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"slug": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\-\s_\.]*$`),
						"must start with a letter or number and may contain spaces, hyphens, underscores, and periods",
					),
				},
				Description: "Repository name",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Repository description",
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"readme": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (receiver *RepositoryResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *RepositoryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		plan  RepositoryModel
		diags diag.Diagnostics
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Create(plan.Project.ValueString(), bitbucket.CreateRepo{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if util.TestError(&response.Diagnostics, err, errorFailedToCreateRepository) {
		return
	}

	plan.Slug = types.StringValue(repository.Slug)

	diags = response.State.SetAttribute(ctx, path.Root("id"), types.StringValue(strconv.Itoa(repository.ID)))
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateState) {
		return
	}

	if plan.Readme.IsNull() {
		_, err = receiver.client.RepositoryService().Initialize(
			plan.Project.ValueString(),
			plan.Name.ValueString(),
			plan.Readme.ValueString(),
		)
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}
	}

	computation, diags := CreateRepositoryAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryModel(repository, plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Read(state.Project.ValueString(), state.Slug.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToCreateRepository) {
		return
	}

	computation, diags := ComputeRepositoryAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryModel(repository, state, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan, state RepositoryModel
	)

	if util.TestDiagnostics(&response.Diagnostics,
		request.Plan.Get(ctx, &plan),
		request.State.Get(ctx, &state)) {
		return
	}

	repository, err := receiver.client.RepositoryService().Update(
		state.Project.ValueString(),
		state.Slug.ValueString(),
		plan.Description.ValueString(),
	)
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
		return
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateRepositoryAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryModel(repository, plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		diags = DeleteRepositoryAssignments(ctx, receiver, state)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}

		err = receiver.client.RepositoryService().Delete(
			state.Project.ValueString(),
			state.Slug.ValueString(),
		)
		if util.TestError(&response.Diagnostics, err, errorFailedToDeleteRepository) {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *RepositoryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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
