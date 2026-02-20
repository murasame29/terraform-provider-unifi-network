// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ datasource.DataSource = &NetworkDataSource{}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

type NetworkDataSource struct {
	client *network.Client
}

type NetworkDataSourceModel struct {
	SiteID                types.String `tfsdk:"site_id"`
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	VlanID                types.Int64  `tfsdk:"vlan_id"`
	Management            types.String `tfsdk:"management"`
	Default               types.Bool   `tfsdk:"default"`
	IsolationEnabled      types.Bool   `tfsdk:"isolation_enabled"`
	InternetAccessEnabled types.Bool   `tfsdk:"internet_access_enabled"`
}

func (d *NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches details of a specific UniFi network.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID where the network is located.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the network.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the network.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the network is enabled.",
				Computed:            true,
			},
			"vlan_id": schema.Int64Attribute{
				MarkdownDescription: "The VLAN ID of the network.",
				Computed:            true,
			},
			"management": schema.StringAttribute{
				MarkdownDescription: "The management type of the network.",
				Computed:            true,
			},
			"default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the default network.",
				Computed:            true,
			},
			"isolation_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether network isolation is enabled.",
				Computed:            true,
			},
			"internet_access_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether internet access is enabled.",
				Computed:            true,
			},
		},
	}
}

func (d *NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*UnifiClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *UnifiClients, got: %T", req.ProviderData),
		)
		return
	}

	d.client = clients.Network
}

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading UniFi network", map[string]interface{}{
		"site_id":    data.SiteID.ValueString(),
		"network_id": data.ID.ValueString(),
	})

	networkResp, err := d.client.GetNetworkDetails(ctx, networktypes.GetNetworkDetailsRequest{
		SiteID:    data.SiteID.ValueString(),
		NetworkID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network: %s", err))
		return
	}

	data.Name = types.StringValue(networkResp.Name)
	data.Enabled = types.BoolValue(networkResp.Enabled)
	data.VlanID = types.Int64Value(int64(networkResp.VlanID))
	data.Management = types.StringValue(networkResp.Management)
	data.Default = types.BoolValue(networkResp.Default)

	if networkResp.IsolationEnabled != nil {
		data.IsolationEnabled = types.BoolValue(*networkResp.IsolationEnabled)
	} else {
		data.IsolationEnabled = types.BoolNull()
	}

	if networkResp.InternetAccessEnabled != nil {
		data.InternetAccessEnabled = types.BoolValue(*networkResp.InternetAccessEnabled)
	} else {
		data.InternetAccessEnabled = types.BoolNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
