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

type RepositoryData struct {
	Project string     `tfsdk:"project"`
	Slug    string     `tfsdk:"slug"`
	Exists  types.Bool `tfsdk:"exists"`
}

var (
	_ datasource.DataSource              = &RepositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoryDataSource{}
	_ ConfigurableReceiver               = &RepositoryDataSource{}
)

func NewRepositoryDataSource() datasource.DataSource {
	return &RepositoryDataSource{}
}

type RepositoryDataSource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryDataSource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *RepositoryDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository"
}

func (receiver *RepositoryDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required: true,
			},
			"slug": schema.StringAttribute{
				Required: true,
			},
			"exists": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (receiver *RepositoryDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		config RepositoryData
	)

	diags = request.Config.Get(ctx, &config)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	_, err = receiver.client.RepositoryService().Read(config.Project, config.Slug)

	diags = response.State.Set(ctx, &RepositoryData{
		Project: config.Project,
		Slug:    config.Slug,
		Exists:  types.BoolValue(err == nil),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
