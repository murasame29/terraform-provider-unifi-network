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

var _ datasource.DataSource = &TrafficMatchingListsDataSource{}

func NewTrafficMatchingListsDataSource() datasource.DataSource {
	return &TrafficMatchingListsDataSource{}
}

type TrafficMatchingListsDataSource struct {
	client *network.Client
}

type TrafficMatchingListsDataSourceModel struct {
	SiteID types.String                   `tfsdk:"site_id"`
	Lists  []TrafficMatchingListSummary   `tfsdk:"lists"`
}

type TrafficMatchingListSummary struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

func (d *TrafficMatchingListsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_matching_lists"
}

func (d *TrafficMatchingListsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of traffic matching lists for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"lists": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
						"type": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *TrafficMatchingListsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TrafficMatchingListsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TrafficMatchingListsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListTrafficMatchingLists(ctx, networktypes.ListTrafficMatchingListsRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read traffic matching lists: %s", err))
		return
	}

	data.Lists = make([]TrafficMatchingListSummary, 0, len(result.Data))
	for _, l := range result.Data {
		data.Lists = append(data.Lists, TrafficMatchingListSummary{
			ID:   types.StringValue(l.ID),
			Name: types.StringValue(l.Name),
			Type: types.StringValue(l.Type),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
