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

var _ datasource.DataSource = &ClientsDataSource{}

func NewClientsDataSource() datasource.DataSource {
	return &ClientsDataSource{}
}

type ClientsDataSource struct {
	client *network.Client
}

type ClientsDataSourceModel struct {
	SiteID  types.String         `tfsdk:"site_id"`
	Clients []ClientSummaryModel `tfsdk:"clients"`
}

type ClientSummaryModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	MacAddress types.String `tfsdk:"mac_address"`
	IPAddress  types.String `tfsdk:"ip_address"`
}

func (d *ClientsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

func (d *ClientsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of connected clients for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"clients": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"type":        schema.StringAttribute{Computed: true},
						"mac_address": schema.StringAttribute{Computed: true},
						"ip_address":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ClientsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClientsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListConnectedClients(ctx, networktypes.ListConnectedClientsRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read clients: %s", err))
		return
	}

	data.Clients = make([]ClientSummaryModel, 0, len(result.Data))
	for _, c := range result.Data {
		data.Clients = append(data.Clients, ClientSummaryModel{
			ID:         types.StringValue(c.ID),
			Name:       types.StringValue(c.Name),
			Type:       types.StringValue(c.Type),
			MacAddress: types.StringValue(c.MacAddress),
			IPAddress:  types.StringValue(c.IPAddress),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
