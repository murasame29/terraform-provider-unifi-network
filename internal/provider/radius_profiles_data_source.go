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

var _ datasource.DataSource = &RadiusProfilesDataSource{}

func NewRadiusProfilesDataSource() datasource.DataSource {
	return &RadiusProfilesDataSource{}
}

type RadiusProfilesDataSource struct {
	client *network.Client
}

type RadiusProfilesDataSourceModel struct {
	SiteID   types.String           `tfsdk:"site_id"`
	Profiles []RadiusProfileSummary `tfsdk:"profiles"`
}

type RadiusProfileSummary struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *RadiusProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radius_profiles"
}

func (d *RadiusProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of RADIUS profiles for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"profiles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *RadiusProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RadiusProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RadiusProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListRadiusProfiles(ctx, networktypes.ListRadiusProfilesRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read RADIUS profiles: %s", err))
		return
	}

	data.Profiles = make([]RadiusProfileSummary, 0, len(result.Data))
	for _, p := range result.Data {
		data.Profiles = append(data.Profiles, RadiusProfileSummary{
			ID:   types.StringValue(p.ID),
			Name: types.StringValue(p.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
