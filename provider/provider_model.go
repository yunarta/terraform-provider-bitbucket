package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
)

type EndPoint struct {
	EndPoint string       `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
}

type Author struct {
	Name  string `tfsdk:"name"`
	Email string `tfsdk:"email"`
}

type BitbucketProviderConfig struct {
	Bitbucket EndPoint `tfsdk:"bitbucket"`
	Author    Author   `tfsdk:"author"`
}

type BitbucketProviderData struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}
