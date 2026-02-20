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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ resource.Resource = &FirewallPolicyResource{}
var _ resource.ResourceWithImportState = &FirewallPolicyResource{}

func NewFirewallPolicyResource() resource.Resource {
	return &FirewallPolicyResource{}
}

type FirewallPolicyResource struct {
	client *network.Client
}

type FirewallPolicyResourceModel struct {
	SiteID            types.String `tfsdk:"site_id"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	ActionType        types.String `tfsdk:"action_type"`
	SourceZoneID      types.String `tfsdk:"source_zone_id"`
	DestinationZoneID types.String `tfsdk:"destination_zone_id"`
	LoggingEnabled    types.Bool   `tfsdk:"logging_enabled"`
}

func (r *FirewallPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_policy"
}

func (r *FirewallPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi firewall policy.",
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
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the firewall policy.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the policy is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"action_type": schema.StringAttribute{
				MarkdownDescription: "The action type (allow/drop/reject). Defaults to `allow`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("allow"),
			},
			"source_zone_id": schema.StringAttribute{
				MarkdownDescription: "The source firewall zone ID.",
				Required:            true,
			},
			"destination_zone_id": schema.StringAttribute{
				MarkdownDescription: "The destination firewall zone ID.",
				Required:            true,
			},
			"logging_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether logging is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *FirewallPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating firewall policy", map[string]interface{}{"name": data.Name.ValueString()})

	createReq := networktypes.CreateFirewallPolicyRequest{
		SiteID:         data.SiteID.ValueString(),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		LoggingEnabled: data.LoggingEnabled.ValueBool(),
		Action: &networktypes.FirewallPolicyAction{
			Type: data.ActionType.ValueString(),
		},
		Source: &networktypes.FirewallPolicyEndpoint{
			ZoneID: data.SourceZoneID.ValueString(),
		},
		Destination: &networktypes.FirewallPolicyEndpoint{
			ZoneID: data.DestinationZoneID.ValueString(),
		},
		IPProtocolScope: &networktypes.FirewallIPProtocolScope{
			IPVersion: "ipv4",
		},
	}

	result, err := r.client.CreateFirewallPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create firewall policy: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetFirewallPolicy(ctx, networktypes.GetFirewallPolicyRequest{
		SiteID:   data.SiteID.ValueString(),
		PolicyID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read firewall policy: %s", err))
		return
	}

	data.Name = types.StringValue(result.Name)
	data.Description = types.StringValue(result.Description)
	data.Enabled = types.BoolValue(result.Enabled)
	data.LoggingEnabled = types.BoolValue(result.LoggingEnabled)
	if result.Action != nil {
		data.ActionType = types.StringValue(result.Action.Type)
	}
	if result.Source != nil {
		data.SourceZoneID = types.StringValue(result.Source.ZoneID)
	}
	if result.Destination != nil {
		data.DestinationZoneID = types.StringValue(result.Destination.ZoneID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := networktypes.UpdateFirewallPolicyRequest{
		SiteID:         data.SiteID.ValueString(),
		PolicyID:       data.ID.ValueString(),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		LoggingEnabled: data.LoggingEnabled.ValueBool(),
		Action: &networktypes.FirewallPolicyAction{
			Type: data.ActionType.ValueString(),
		},
		Source: &networktypes.FirewallPolicyEndpoint{
			ZoneID: data.SourceZoneID.ValueString(),
		},
		Destination: &networktypes.FirewallPolicyEndpoint{
			ZoneID: data.DestinationZoneID.ValueString(),
		},
		IPProtocolScope: &networktypes.FirewallIPProtocolScope{
			IPVersion: "ipv4",
		},
	}

	_, err := r.client.UpdateFirewallPolicy(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update firewall policy: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFirewallPolicy(ctx, networktypes.DeleteFirewallPolicyRequest{
		SiteID:   data.SiteID.ValueString(),
		PolicyID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete firewall policy: %s", err))
		return
	}
}

func (r *FirewallPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
