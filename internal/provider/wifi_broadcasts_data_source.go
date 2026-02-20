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

var _ datasource.DataSource = &WifiBroadcastsDataSource{}

func NewWifiBroadcastsDataSource() datasource.DataSource {
	return &WifiBroadcastsDataSource{}
}

type WifiBroadcastsDataSource struct {
	client *network.Client
}

type WifiBroadcastsDataSourceModel struct {
	SiteID     types.String             `tfsdk:"site_id"`
	Broadcasts []WifiBroadcastSummary   `tfsdk:"broadcasts"`
}

type WifiBroadcastSummary struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (d *WifiBroadcastsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wifi_broadcasts"
}

func (d *WifiBroadcastsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of WiFi broadcasts (SSIDs) for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"broadcasts": schema.ListNestedAttribute{
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

func (d *WifiBroadcastsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WifiBroadcastsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WifiBroadcastsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListWifiBroadcasts(ctx, networktypes.ListWifiBroadcastsRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read WiFi broadcasts: %s", err))
		return
	}

	data.Broadcasts = make([]WifiBroadcastSummary, 0, len(result.Data))
	for _, b := range result.Data {
		data.Broadcasts = append(data.Broadcasts, WifiBroadcastSummary{
			ID:      types.StringValue(b.ID),
			Name:    types.StringValue(b.Name),
			Type:    types.StringValue(b.Type),
			Enabled: types.BoolValue(b.Enabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
