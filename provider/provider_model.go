package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type EndPoint struct {
	EndPoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
}

type BitbucketProviderConfig struct {
	Bitbucket EndPoint `tfsdk:"bitbucket"`
}

type BitbucketProviderData struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}
