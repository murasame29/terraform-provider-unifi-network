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

var _ datasource.DataSource = &SitesDataSource{}

func NewSitesDataSource() datasource.DataSource {
	return &SitesDataSource{}
}

type SitesDataSource struct {
	client *network.Client
}

type SitesDataSourceModel struct {
	Sites []SiteModel `tfsdk:"sites"`
}

type SiteModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	InternalReference types.String `tfsdk:"internal_reference"`
}

func (d *SitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

func (d *SitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of UniFi sites.",
		Attributes: map[string]schema.Attribute{
			"sites": schema.ListNestedAttribute{
				MarkdownDescription: "List of UniFi sites.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the site.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the site.",
							Computed:            true,
						},
						"internal_reference": schema.StringAttribute{
							MarkdownDescription: "The internal reference of the site.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SitesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SitesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading UniFi sites")

	sitesResp, err := d.client.ListSites(ctx, networktypes.ListSitesRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sites: %s", err))
		return
	}

	data.Sites = make([]SiteModel, 0, len(sitesResp.Data))
	for _, site := range sitesResp.Data {
		data.Sites = append(data.Sites, SiteModel{
			ID:                types.StringValue(site.ID),
			Name:              types.StringValue(site.Name),
			InternalReference: types.StringValue(site.InternalReference),
		})
	}

	tflog.Debug(ctx, fmt.Sprintf("Read %d sites", len(data.Sites)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
