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

var _ datasource.DataSource = &NetworksDataSource{}

func NewNetworksDataSource() datasource.DataSource {
	return &NetworksDataSource{}
}

type NetworksDataSource struct {
	client *network.Client
}

type NetworksDataSourceModel struct {
	SiteID   types.String          `tfsdk:"site_id"`
	Networks []NetworkSummaryModel `tfsdk:"networks"`
}

type NetworkSummaryModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	VlanID     types.Int64  `tfsdk:"vlan_id"`
	Management types.String `tfsdk:"management"`
	Default    types.Bool   `tfsdk:"default"`
}

func (d *NetworksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *NetworksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of networks for a UniFi site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID to list networks for.",
				Required:            true,
			},
			"networks": schema.ListNestedAttribute{
				MarkdownDescription: "List of networks.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the network.",
							Computed:            true,
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
					},
				},
			},
		},
	}
}

func (d *NetworksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworksDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading UniFi networks", map[string]interface{}{
		"site_id": data.SiteID.ValueString(),
	})

	networksResp, err := d.client.ListNetworks(ctx, networktypes.ListNetworksRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read networks: %s", err))
		return
	}

	data.Networks = make([]NetworkSummaryModel, 0, len(networksResp.Data))
	for _, n := range networksResp.Data {
		data.Networks = append(data.Networks, NetworkSummaryModel{
			ID:         types.StringValue(n.ID),
			Name:       types.StringValue(n.Name),
			Enabled:    types.BoolValue(n.Enabled),
			VlanID:     types.Int64Value(int64(n.VlanID)),
			Management: types.StringValue(n.Management),
			Default:    types.BoolValue(n.Default),
		})
	}

	tflog.Debug(ctx, fmt.Sprintf("Read %d networks", len(data.Networks)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
