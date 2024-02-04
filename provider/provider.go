package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yunarta/terraform-api-transport/transport"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
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
					"username": schema.StringAttribute{
						Optional:  true,
						Sensitive: false,
					},
					"password": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},
					"token": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},
				},
			},
			"author": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required: true,
					},
					"email": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (p *BitbucketProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config BitbucketProviderConfig

	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if token is provided
	if !config.Bitbucket.Token.IsNull() {
		// If token is provided, username and password should not be set
		if !config.Bitbucket.Username.IsNull() || !config.Bitbucket.Password.IsNull() {
			response.Diagnostics.AddError(
				"Invalid Configuration",
				"When 'token' is provided, 'username' and 'password' must not be set.",
			)
			return
		}
	} else {
		// If token is not provided, both username and password are required
		if config.Bitbucket.Username.IsNull() || config.Bitbucket.Password.IsNull() {
			response.Diagnostics.AddError(
				"Invalid Configuration",
				"Both 'username' and 'password' must be set when 'token' is not provided.",
			)
			return
		}
	}

	var authentication transport.Authentication
	if !config.Bitbucket.Token.IsNull() {
		authentication = transport.BearerAuthentication{
			Token: config.Bitbucket.Token.ValueString(),
		}
	} else {
		authentication = transport.BasicAuthentication{
			Username: config.Bitbucket.Username.ValueString(),
			Password: config.Bitbucket.Password.ValueString(),
		}
	}

	providerData := &BitbucketProviderData{
		config: config,
		client: bitbucket.NewBitbucketClient(
			transport.NewHttpPayloadTransport(config.Bitbucket.EndPoint,
				authentication,
			),
		),
	}

	response.DataSourceData = providerData
	response.ResourceData = providerData
}

func (p *BitbucketProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRepositoryPermissionsDataSource,
		NewProjectPermissionsDataSource,
	}
}

func (p *BitbucketProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewProjectPermissionsResource,
		NewProjectBranchRestrictionsResource,
		NewProjectMergeChecksResource,
		NewRepositoryResource,
		NewRepositoryPermissionsResource,
		NewRepositoryBranchRestrictionsResource,
		NewRepositoryMergeChecksResource,
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
