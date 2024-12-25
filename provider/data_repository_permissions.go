package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
)

type RepositoryPermissionsData struct {
	Project string              `tfsdk:"project"`
	Slug    string              `tfsdk:"slug"`
	Users   map[string][]string `tfsdk:"users"`
	Groups  map[string][]string `tfsdk:"groups"`
}

var (
	_ datasource.DataSource              = &RepositoryPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoryPermissionsDataSource{}
	_ ConfigurableReceiver               = &RepositoryPermissionsDataSource{}
)

func NewRepositoryPermissionsDataSource() datasource.DataSource {
	return &RepositoryPermissionsDataSource{}
}

type RepositoryPermissionsDataSource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryPermissionsDataSource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryPermissionsDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *RepositoryPermissionsDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository_permissions"
}

func (receiver *RepositoryPermissionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required: true,
			},
			"slug": schema.StringAttribute{
				Required: true,
			},
			"users": schema.MapAttribute{
				Computed: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
			},
			"groups": schema.MapAttribute{
				Computed: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
			},
		},
	}
}

func (receiver *RepositoryPermissionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		config RepositoryPermissionsData
		diags  diag.Diagnostics
	)

	diags = request.Config.Get(ctx, &config)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	permissions, err := receiver.client.RepositoryService().ReadPermissions(config.Project, config.Slug)
	if util.TestError(&response.Diagnostics, err, "") {
		return
	}

	if permissions == nil {
		response.Diagnostics.AddError("Unable to find deployment", config.Project)
		return
	}

	users, groups := CreateAttestation(permissions, []string{"REPO_ADMIN", "REPO_READ", "REPO_WRITE"})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &RepositoryPermissionsData{
		Project: config.Project,
		Users:   users,
		Groups:  groups,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
