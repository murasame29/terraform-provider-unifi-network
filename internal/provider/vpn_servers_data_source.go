// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ datasource.DataSource = &VPNServersDataSource{}

func NewVPNServersDataSource() datasource.DataSource {
	return &VPNServersDataSource{}
}

type VPNServersDataSource struct {
	client *network.Client
}

type VPNServersDataSourceModel struct {
	SiteID  types.String       `tfsdk:"site_id"`
	Servers []VPNServerSummary `tfsdk:"servers"`
}

type VPNServerSummary struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (d *VPNServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_servers"
}

func (d *VPNServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of VPN servers for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"servers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"type":    schema.StringAttribute{Computed: true},
						"enabled": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *VPNServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clients, ok := req.ProviderData.(*UnifiClients)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *UnifiClients, got: %T", req.ProviderData))
		return
	}
	d.client = clients.Network
}

func (d *VPNServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VPNServersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListVPNServers(ctx, networktypes.ListVPNServersRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VPN servers: %s", err))
		return
	}

	data.Servers = make([]VPNServerSummary, 0, len(result.Data))
	for _, s := range result.Data {
		data.Servers = append(data.Servers, VPNServerSummary{
			ID:      types.StringValue(s.ID),
			Name:    types.StringValue(s.Name),
			Type:    types.StringValue(s.Type),
			Enabled: types.BoolValue(s.Enabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
