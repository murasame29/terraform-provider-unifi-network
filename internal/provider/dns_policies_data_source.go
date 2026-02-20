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

var _ datasource.DataSource = &DNSPoliciesDataSource{}

func NewDNSPoliciesDataSource() datasource.DataSource {
	return &DNSPoliciesDataSource{}
}

type DNSPoliciesDataSource struct {
	client *network.Client
}

type DNSPoliciesDataSourceModel struct {
	SiteID   types.String       `tfsdk:"site_id"`
	Policies []DNSPolicySummary `tfsdk:"policies"`
}

type DNSPolicySummary struct {
	ID      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Domain  types.String `tfsdk:"domain"`
}

func (d *DNSPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_policies"
}

func (d *DNSPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of DNS policies for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"type":    schema.StringAttribute{Computed: true},
						"enabled": schema.BoolAttribute{Computed: true},
						"domain":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *DNSPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListDNSPolicies(ctx, networktypes.ListDNSPoliciesRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS policies: %s", err))
		return
	}

	data.Policies = make([]DNSPolicySummary, 0, len(result.Data))
	for _, p := range result.Data {
		data.Policies = append(data.Policies, DNSPolicySummary{
			ID:      types.StringValue(p.ID),
			Type:    types.StringValue(p.Type),
			Enabled: types.BoolValue(p.Enabled),
			Domain:  types.StringValue(p.Domain),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
