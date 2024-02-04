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

type ProjectPermissionsData struct {
	Key    string              `tfsdk:"key"`
	Users  map[string][]string `tfsdk:"users"`
	Groups map[string][]string `tfsdk:"groups"`
}

var (
	_ datasource.DataSource              = &ProjectPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectPermissionsDataSource{}
	_ ConfigurableReceiver               = &ProjectPermissionsDataSource{}
)

func NewProjectPermissionsDataSource() datasource.DataSource {
	return &ProjectPermissionsDataSource{}
}

type ProjectPermissionsDataSource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *ProjectPermissionsDataSource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectPermissionsDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *ProjectPermissionsDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_permissions"
}

func (receiver *ProjectPermissionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
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

func (receiver *ProjectPermissionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		config ProjectPermissionsData
		diags  diag.Diagnostics
	)

	diags = request.Config.Get(ctx, &config)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	permissions, err := receiver.client.ProjectService().ReadPermissions(config.Key)
	if util.TestError(&response.Diagnostics, err, "") {
		return
	}

	if permissions == nil {
		response.Diagnostics.AddError("Unable to find deployment", config.Key)
		return
	}

	users, groups := CreateAttestation(permissions, []string{"PROJECT_ADMIN", "REPO_CREATE", "PROJECT_READ", "PROJECT_WRITE"})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &ProjectPermissionsData{
		Key:    config.Key,
		Users:  users,
		Groups: groups,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
