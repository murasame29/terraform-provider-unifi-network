// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ resource.Resource = &DNSPolicyResource{}
var _ resource.ResourceWithImportState = &DNSPolicyResource{}

func NewDNSPolicyResource() resource.Resource {
	return &DNSPolicyResource{}
}

type DNSPolicyResource struct {
	client *network.Client
}

type DNSPolicyResourceModel struct {
	SiteID      types.String `tfsdk:"site_id"`
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Domain      types.String `tfsdk:"domain"`
	IPv4Address types.String `tfsdk:"ipv4_address"`
	IPv6Address types.String `tfsdk:"ipv6_address"`
	TTLSeconds  types.Int64  `tfsdk:"ttl_seconds"`
}

func (r *DNSPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_policy"
}

func (r *DNSPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi DNS policy.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The DNS record type (A, AAAA, CNAME, etc.).",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the policy is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Optional:            true,
			},
			"ipv4_address": schema.StringAttribute{
				MarkdownDescription: "The IPv4 address (for A records).",
				Optional:            true,
			},
			"ipv6_address": schema.StringAttribute{
				MarkdownDescription: "The IPv6 address (for AAAA records).",
				Optional:            true,
			},
			"ttl_seconds": schema.Int64Attribute{
				MarkdownDescription: "The TTL in seconds.",
				Optional:            true,
			},
		},
	}
}

func (r *DNSPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DNS policy", map[string]interface{}{"type": data.Type.ValueString()})

	createReq := networktypes.CreateDNSPolicyRequest{
		SiteID:      data.SiteID.ValueString(),
		Type:        data.Type.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Domain:      data.Domain.ValueString(),
		IPv4Address: data.IPv4Address.ValueString(),
		IPv6Address: data.IPv6Address.ValueString(),
	}

	if !data.TTLSeconds.IsNull() {
		ttl := int(data.TTLSeconds.ValueInt64())
		createReq.TTLSeconds = &ttl
	}

	result, err := r.client.CreateDNSPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create DNS policy: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetDNSPolicy(ctx, networktypes.GetDNSPolicyRequest{
		SiteID:   data.SiteID.ValueString(),
		PolicyID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS policy: %s", err))
		return
	}

	data.Type = types.StringValue(result.Type)
	data.Enabled = types.BoolValue(result.Enabled)
	data.Domain = types.StringValue(result.Domain)
	data.IPv4Address = types.StringValue(result.IPv4Address)
	data.IPv6Address = types.StringValue(result.IPv6Address)
	if result.TTLSeconds != nil {
		data.TTLSeconds = types.Int64Value(int64(*result.TTLSeconds))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := networktypes.UpdateDNSPolicyRequest{
		SiteID:      data.SiteID.ValueString(),
		PolicyID:    data.ID.ValueString(),
		Type:        data.Type.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Domain:      data.Domain.ValueString(),
		IPv4Address: data.IPv4Address.ValueString(),
		IPv6Address: data.IPv6Address.ValueString(),
	}

	if !data.TTLSeconds.IsNull() {
		ttl := int(data.TTLSeconds.ValueInt64())
		updateReq.TTLSeconds = &ttl
	}

	_, err := r.client.UpdateDNSPolicy(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update DNS policy: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSPolicy(ctx, networktypes.DeleteDNSPolicyRequest{
		SiteID:   data.SiteID.ValueString(),
		PolicyID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNS policy: %s", err))
		return
	}
}

func (r *DNSPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
