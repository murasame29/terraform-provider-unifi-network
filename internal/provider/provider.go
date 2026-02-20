// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	sitemanager "github.com/murasame29/unifi-client-go/services/site-manager"
)

var _ provider.Provider = &UnifiNetworkProvider{}

type UnifiNetworkProvider struct {
	version string
}

type UnifiNetworkProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

type UnifiClients struct {
	Network     *network.Client
	SiteManager *sitemanager.Client
}

func (p *UnifiNetworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unifi"
	resp.Version = p.version
}

func (p *UnifiNetworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The UniFi Network provider allows you to manage UniFi Network resources using the UniFi Cloud API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authenticating with the UniFi Cloud API. Can also be set via the `UNIFI_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL for the UniFi Cloud API. Defaults to `https://api.ui.com`. Can also be set via the `UNIFI_BASE_URL` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *UnifiNetworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring UniFi Network provider")

	var config UnifiNetworkProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("UNIFI_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing UniFi API Key",
			"The provider cannot create the UniFi API client as there is a missing or empty value for the UniFi API key. "+
				"Set the api_key value in the configuration or use the UNIFI_API_KEY environment variable.",
		)
		return
	}

	baseURL := os.Getenv("UNIFI_BASE_URL")
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	var opts []network.Option
	if baseURL != "" {
		opts = append(opts, network.WithBaseURL(baseURL))
	}

	clients := &UnifiClients{
		Network:     network.NewClient(apiKey, opts...),
		SiteManager: sitemanager.NewClient(apiKey, opts...),
	}

	tflog.Debug(ctx, "Created UniFi API clients")

	resp.DataSourceData = clients
	resp.ResourceData = clients
}

func (p *UnifiNetworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNetworkResource,
		NewWifiBroadcastResource,
		NewACLRuleResource,
		NewDNSPolicyResource,
		NewFirewallZoneResource,
		NewFirewallPolicyResource,
		NewTrafficMatchingListResource,
		NewVoucherResource,
	}
}

func (p *UnifiNetworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSitesDataSource,
		NewNetworkDataSource,
		NewNetworksDataSource,
		NewDevicesDataSource,
		NewDeviceDataSource,
		NewClientsDataSource,
		NewACLRulesDataSource,
		NewDNSPoliciesDataSource,
		NewFirewallZonesDataSource,
		NewFirewallPoliciesDataSource,
		NewTrafficMatchingListsDataSource,
		NewVouchersDataSource,
		NewWANInterfacesDataSource,
		NewVPNTunnelsDataSource,
		NewVPNServersDataSource,
		NewRadiusProfilesDataSource,
		NewWifiBroadcastsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnifiNetworkProvider{
			version: version,
		}
	}
}
