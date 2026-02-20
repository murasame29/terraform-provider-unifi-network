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

var _ datasource.DataSource = &VouchersDataSource{}

func NewVouchersDataSource() datasource.DataSource {
	return &VouchersDataSource{}
}

type VouchersDataSource struct {
	client *network.Client
}

type VouchersDataSourceModel struct {
	SiteID   types.String       `tfsdk:"site_id"`
	Vouchers []VoucherSummary   `tfsdk:"vouchers"`
}

type VoucherSummary struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Code             types.String `tfsdk:"code"`
	TimeLimitMinutes types.Int64  `tfsdk:"time_limit_minutes"`
	Expired          types.Bool   `tfsdk:"expired"`
}

func (d *VouchersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vouchers"
}

func (d *VouchersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of hotspot vouchers for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{Required: true},
			"vouchers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 schema.StringAttribute{Computed: true},
						"name":               schema.StringAttribute{Computed: true},
						"code":               schema.StringAttribute{Computed: true},
						"time_limit_minutes": schema.Int64Attribute{Computed: true},
						"expired":            schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *VouchersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VouchersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VouchersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListVouchers(ctx, networktypes.ListVouchersRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vouchers: %s", err))
		return
	}

	data.Vouchers = make([]VoucherSummary, 0, len(result.Data))
	for _, v := range result.Data {
		data.Vouchers = append(data.Vouchers, VoucherSummary{
			ID:               types.StringValue(v.ID),
			Name:             types.StringValue(v.Name),
			Code:             types.StringValue(v.Code),
			TimeLimitMinutes: types.Int64Value(int64(v.TimeLimitMinutes)),
			Expired:          types.BoolValue(v.Expired),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
