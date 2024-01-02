package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type EndPoint struct {
	EndPoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

type BitbucketProviderConfig struct {
	Bitbucket EndPoint `tfsdk:"bitbucket"`
}
