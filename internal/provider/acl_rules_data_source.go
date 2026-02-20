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

var _ datasource.DataSource = &ACLRulesDataSource{}

func NewACLRulesDataSource() datasource.DataSource {
	return &ACLRulesDataSource{}
}

type ACLRulesDataSource struct {
	client *network.Client
}

type ACLRulesDataSourceModel struct {
	SiteID types.String     `tfsdk:"site_id"`
	Rules  []ACLRuleSummary `tfsdk:"rules"`
}

type ACLRuleSummary struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Action  types.String `tfsdk:"action"`
}

func (d *ACLRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_rules"
}

func (d *ACLRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of ACL rules for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"type":    schema.StringAttribute{Computed: true},
						"enabled": schema.BoolAttribute{Computed: true},
						"action":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ACLRulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ACLRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ACLRulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListACLRules(ctx, networktypes.ListACLRulesRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ACL rules: %s", err))
		return
	}

	data.Rules = make([]ACLRuleSummary, 0, len(result.Data))
	for _, r := range result.Data {
		data.Rules = append(data.Rules, ACLRuleSummary{
			ID:      types.StringValue(r.ID),
			Name:    types.StringValue(r.Name),
			Type:    types.StringValue(r.Type),
			Enabled: types.BoolValue(r.Enabled),
			Action:  types.StringValue(r.Action),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
