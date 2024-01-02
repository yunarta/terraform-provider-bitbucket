package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type BitbucketProvider struct {
	Version string
}

func (p *BitbucketProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "bitbucket"
	response.Version = p.Version
}

func (p *BitbucketProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"bitbucket": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required: true,
					},
					"token": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
				},
			},
		},
	}
}

func (p *BitbucketProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config *BitbucketProviderConfig

	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.DataSourceData = config
	response.ResourceData = config
}

func (p *BitbucketProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *BitbucketProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRepositoryResource,
		NewProjectPermissionsResource,
		NewRepositoryPermissionsResource,
	}
}

var _ provider.Provider = &BitbucketProvider{}

func New(Version string) func() provider.Provider {
	return func() provider.Provider {
		return &BitbucketProvider{
			Version: Version,
		}
	}
}
