// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ resource.Resource = &VoucherResource{}

func NewVoucherResource() resource.Resource {
	return &VoucherResource{}
}

type VoucherResource struct {
	client *network.Client
}

type VoucherResourceModel struct {
	SiteID               types.String `tfsdk:"site_id"`
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Code                 types.String `tfsdk:"code"`
	TimeLimitMinutes     types.Int64  `tfsdk:"time_limit_minutes"`
	VoucherCount         types.Int64  `tfsdk:"voucher_count"`
	AuthorizedGuestLimit types.Int64  `tfsdk:"authorized_guest_limit"`
	DataUsageLimitMBytes types.Int64  `tfsdk:"data_usage_limit_mbytes"`
	RxRateLimitKbps      types.Int64  `tfsdk:"rx_rate_limit_kbps"`
	TxRateLimitKbps      types.Int64  `tfsdk:"tx_rate_limit_kbps"`
}

func (r *VoucherResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voucher"
}

func (r *VoucherResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages UniFi hotspot vouchers for guest access.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the first voucher.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name/note for the voucher.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"code": schema.StringAttribute{
				MarkdownDescription: "The voucher code (generated).",
				Computed:            true,
			},
			"time_limit_minutes": schema.Int64Attribute{
				MarkdownDescription: "Time limit in minutes. Defaults to `60`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
			},
			"voucher_count": schema.Int64Attribute{
				MarkdownDescription: "Number of vouchers to generate. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"authorized_guest_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of guests that can use this voucher. Leave empty for unlimited.",
				Optional:            true,
			},
			"data_usage_limit_mbytes": schema.Int64Attribute{
				MarkdownDescription: "Data usage limit in megabytes. Leave empty for unlimited.",
				Optional:            true,
			},
			"rx_rate_limit_kbps": schema.Int64Attribute{
				MarkdownDescription: "Download rate limit in kbps. Leave empty for unlimited.",
				Optional:            true,
			},
			"tx_rate_limit_kbps": schema.Int64Attribute{
				MarkdownDescription: "Upload rate limit in kbps. Leave empty for unlimited.",
				Optional:            true,
			},
		},
	}
}

func (r *VoucherResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clients, ok := req.ProviderData.(*UnifiClients)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *UnifiClients, got: %T", req.ProviderData))
		return
	}
	r.client = clients.Network
}

func (r *VoucherResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VoucherResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	count := int(data.VoucherCount.ValueInt64())
	createReq := networktypes.GenerateVouchersRequest{
		SiteID:           data.SiteID.ValueString(),
		Name:             data.Name.ValueString(),
		TimeLimitMinutes: int(data.TimeLimitMinutes.ValueInt64()),
		Count:            &count,
	}

	if !data.AuthorizedGuestLimit.IsNull() {
		limit := int(data.AuthorizedGuestLimit.ValueInt64())
		createReq.AuthorizedGuestLimit = &limit
	}
	if !data.DataUsageLimitMBytes.IsNull() {
		limit := int(data.DataUsageLimitMBytes.ValueInt64())
		createReq.DataUsageLimitMBytes = &limit
	}
	if !data.RxRateLimitKbps.IsNull() {
		limit := int(data.RxRateLimitKbps.ValueInt64())
		createReq.RxRateLimitKbps = &limit
	}
	if !data.TxRateLimitKbps.IsNull() {
		limit := int(data.TxRateLimitKbps.ValueInt64())
		createReq.TxRateLimitKbps = &limit
	}

	vouchersResp, err := r.client.GenerateVouchers(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create voucher: %s", err))
		return
	}

	if len(vouchersResp.Vouchers) > 0 {
		data.ID = types.StringValue(vouchersResp.Vouchers[0].ID)
		data.Code = types.StringValue(vouchersResp.Vouchers[0].Code)
	}

	tflog.Trace(ctx, "created voucher resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VoucherResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VoucherResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	voucher, err := r.client.GetVoucherDetails(ctx, networktypes.GetVoucherDetailsRequest{
		SiteID:    data.SiteID.ValueString(),
		VoucherID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read voucher: %s", err))
		return
	}

	data.Name = types.StringValue(voucher.Name)
	data.Code = types.StringValue(voucher.Code)
	data.TimeLimitMinutes = types.Int64Value(int64(voucher.TimeLimitMinutes))
	if voucher.AuthorizedGuestLimit != nil {
		data.AuthorizedGuestLimit = types.Int64Value(int64(*voucher.AuthorizedGuestLimit))
	}
	if voucher.DataUsageLimitMBytes != nil {
		data.DataUsageLimitMBytes = types.Int64Value(int64(*voucher.DataUsageLimitMBytes))
	}
	if voucher.RxRateLimitKbps != nil {
		data.RxRateLimitKbps = types.Int64Value(int64(*voucher.RxRateLimitKbps))
	}
	if voucher.TxRateLimitKbps != nil {
		data.TxRateLimitKbps = types.Int64Value(int64(*voucher.TxRateLimitKbps))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VoucherResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VoucherResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Vouchers are immutable - updates require replacement
	resp.Diagnostics.AddError("Update Not Supported", "Vouchers cannot be updated. Please delete and recreate.")
}

func (r *VoucherResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VoucherResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteVoucher(ctx, networktypes.DeleteVoucherRequest{
		SiteID:    data.SiteID.ValueString(),
		VoucherID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete voucher: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted voucher resource")
}
