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

var _ datasource.DataSource = &FirewallPoliciesDataSource{}

func NewFirewallPoliciesDataSource() datasource.DataSource {
	return &FirewallPoliciesDataSource{}
}

type FirewallPoliciesDataSource struct {
	client *network.Client
}

type FirewallPoliciesDataSourceModel struct {
	SiteID   types.String             `tfsdk:"site_id"`
	Policies []FirewallPolicySummary  `tfsdk:"policies"`
}

type FirewallPolicySummary struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (d *FirewallPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_policies"
}

func (d *FirewallPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of firewall policies for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"enabled": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *FirewallPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FirewallPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListFirewallPolicies(ctx, networktypes.ListFirewallPoliciesRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read firewall policies: %s", err))
		return
	}

	data.Policies = make([]FirewallPolicySummary, 0, len(result.Data))
	for _, p := range result.Data {
		data.Policies = append(data.Policies, FirewallPolicySummary{
			ID:      types.StringValue(p.ID),
			Name:    types.StringValue(p.Name),
			Enabled: types.BoolValue(p.Enabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
