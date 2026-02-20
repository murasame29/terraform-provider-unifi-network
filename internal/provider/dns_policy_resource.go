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
	SiteID           types.String `tfsdk:"site_id"`
	ID               types.String `tfsdk:"id"`
	Type             types.String `tfsdk:"type"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Domain           types.String `tfsdk:"domain"`
	IPv4Address      types.String `tfsdk:"ipv4_address"`
	IPv6Address      types.String `tfsdk:"ipv6_address"`
	TargetDomain     types.String `tfsdk:"target_domain"`
	MailServerDomain types.String `tfsdk:"mail_server_domain"`
	Priority         types.Int64  `tfsdk:"priority"`
	Text             types.String `tfsdk:"text"`
	ServerDomain     types.String `tfsdk:"server_domain"`
	Service          types.String `tfsdk:"service"`
	Protocol         types.String `tfsdk:"protocol"`
	Port             types.Int64  `tfsdk:"port"`
	Weight           types.Int64  `tfsdk:"weight"`
	IPAddress        types.String `tfsdk:"ip_address"`
	TTLSeconds       types.Int64  `tfsdk:"ttl_seconds"`
}

func (r *DNSPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_policy"
}

func (r *DNSPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi DNS policy (local DNS record).",
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
				MarkdownDescription: "The DNS record type (A, AAAA, CNAME, MX, TXT, SRV, PTR).",
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
			"target_domain": schema.StringAttribute{
				MarkdownDescription: "The target domain (for CNAME records).",
				Optional:            true,
			},
			"mail_server_domain": schema.StringAttribute{
				MarkdownDescription: "The mail server domain (for MX records).",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority (for MX and SRV records).",
				Optional:            true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The text content (for TXT records).",
				Optional:            true,
			},
			"server_domain": schema.StringAttribute{
				MarkdownDescription: "The server domain (for SRV records).",
				Optional:            true,
			},
			"service": schema.StringAttribute{
				MarkdownDescription: "The service name (for SRV records, e.g., _sip).",
				Optional:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol (for SRV records, e.g., _tcp, _udp).",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port number (for SRV records).",
				Optional:            true,
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "The weight (for SRV records).",
				Optional:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "The IP address (for PTR records).",
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
		SiteID:           data.SiteID.ValueString(),
		Type:             data.Type.ValueString(),
		Enabled:          data.Enabled.ValueBool(),
		Domain:           data.Domain.ValueString(),
		IPv4Address:      data.IPv4Address.ValueString(),
		IPv6Address:      data.IPv6Address.ValueString(),
		TargetDomain:     data.TargetDomain.ValueString(),
		MailServerDomain: data.MailServerDomain.ValueString(),
		Text:             data.Text.ValueString(),
		ServerDomain:     data.ServerDomain.ValueString(),
		Service:          data.Service.ValueString(),
		Protocol:         data.Protocol.ValueString(),
		IPAddress:        data.IPAddress.ValueString(),
	}

	if !data.Priority.IsNull() {
		priority := int(data.Priority.ValueInt64())
		createReq.Priority = &priority
	}
	if !data.Port.IsNull() {
		port := int(data.Port.ValueInt64())
		createReq.Port = &port
	}
	if !data.Weight.IsNull() {
		weight := int(data.Weight.ValueInt64())
		createReq.Weight = &weight
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
	data.TargetDomain = types.StringValue(result.TargetDomain)
	data.MailServerDomain = types.StringValue(result.MailServerDomain)
	data.Text = types.StringValue(result.Text)
	data.ServerDomain = types.StringValue(result.ServerDomain)
	data.Service = types.StringValue(result.Service)
	data.Protocol = types.StringValue(result.Protocol)
	data.IPAddress = types.StringValue(result.IPAddress)

	if result.Priority != nil {
		data.Priority = types.Int64Value(int64(*result.Priority))
	}
	if result.Port != nil {
		data.Port = types.Int64Value(int64(*result.Port))
	}
	if result.Weight != nil {
		data.Weight = types.Int64Value(int64(*result.Weight))
	}
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
		SiteID:           data.SiteID.ValueString(),
		PolicyID:         data.ID.ValueString(),
		Type:             data.Type.ValueString(),
		Enabled:          data.Enabled.ValueBool(),
		Domain:           data.Domain.ValueString(),
		IPv4Address:      data.IPv4Address.ValueString(),
		IPv6Address:      data.IPv6Address.ValueString(),
		TargetDomain:     data.TargetDomain.ValueString(),
		MailServerDomain: data.MailServerDomain.ValueString(),
		Text:             data.Text.ValueString(),
		ServerDomain:     data.ServerDomain.ValueString(),
		Service:          data.Service.ValueString(),
		Protocol:         data.Protocol.ValueString(),
		IPAddress:        data.IPAddress.ValueString(),
	}

	if !data.Priority.IsNull() {
		priority := int(data.Priority.ValueInt64())
		updateReq.Priority = &priority
	}
	if !data.Port.IsNull() {
		port := int(data.Port.ValueInt64())
		updateReq.Port = &port
	}
	if !data.Weight.IsNull() {
		weight := int(data.Weight.ValueInt64())
		updateReq.Weight = &weight
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
