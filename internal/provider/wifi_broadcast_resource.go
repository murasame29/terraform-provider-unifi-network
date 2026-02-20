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

var _ resource.Resource = &WifiBroadcastResource{}
var _ resource.ResourceWithImportState = &WifiBroadcastResource{}

func NewWifiBroadcastResource() resource.Resource {
	return &WifiBroadcastResource{}
}

type WifiBroadcastResource struct {
	client *network.Client
}

type WifiBroadcastResourceModel struct {
	SiteID                 types.String `tfsdk:"site_id"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Type                   types.String `tfsdk:"type"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	NetworkID              types.String `tfsdk:"network_id"`
	SecurityType           types.String `tfsdk:"security_type"`
	Passphrase             types.String `tfsdk:"passphrase"`
	HideName               types.Bool   `tfsdk:"hide_name"`
	ClientIsolationEnabled types.Bool   `tfsdk:"client_isolation_enabled"`
}

func (r *WifiBroadcastResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wifi_broadcast"
}

func (r *WifiBroadcastResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi WiFi broadcast (SSID).",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID where the WiFi broadcast will be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the WiFi broadcast.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name (SSID) of the WiFi broadcast.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of WiFi broadcast. Defaults to `standard`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("standard"),
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the WiFi broadcast is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The network ID to associate with this WiFi broadcast.",
				Optional:            true,
			},
			"security_type": schema.StringAttribute{
				MarkdownDescription: "The security type. Valid values: `open`, `wpa2`, `wpa3`, `wpa2wpa3`. Defaults to `wpa2`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("wpa2"),
			},
			"passphrase": schema.StringAttribute{
				MarkdownDescription: "The WiFi passphrase. Required when security_type is not `open`.",
				Optional:            true,
				Sensitive:           true,
			},
			"hide_name": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide the SSID. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"client_isolation_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether client isolation is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *WifiBroadcastResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*UnifiClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *UnifiClients, got: %T", req.ProviderData),
		)
		return
	}

	r.client = clients.Network
}

func (r *WifiBroadcastResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WifiBroadcastResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating UniFi WiFi broadcast", map[string]interface{}{
		"site_id": data.SiteID.ValueString(),
		"name":    data.Name.ValueString(),
	})

	createReq := networktypes.CreateWifiBroadcastRequest{
		SiteID:                 data.SiteID.ValueString(),
		Name:                   data.Name.ValueString(),
		Type:                   data.Type.ValueString(),
		Enabled:                data.Enabled.ValueBool(),
		HideName:               data.HideName.ValueBool(),
		ClientIsolationEnabled: data.ClientIsolationEnabled.ValueBool(),
	}

	// Set network reference if provided
	if !data.NetworkID.IsNull() && !data.NetworkID.IsUnknown() {
		createReq.Network = &networktypes.WifiNetworkReference{
			Type:      "network",
			NetworkID: data.NetworkID.ValueString(),
		}
	}

	// Set security configuration
	if !data.SecurityType.IsNull() {
		securityConfig := &networktypes.WifiSecurityConfiguration{
			Type: data.SecurityType.ValueString(),
		}
		if !data.Passphrase.IsNull() && !data.Passphrase.IsUnknown() {
			securityConfig.Passphrase = data.Passphrase.ValueString()
		}
		createReq.SecurityConfiguration = securityConfig
	}

	wifiResp, err := r.client.CreateWifiBroadcast(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create WiFi broadcast: %s", err))
		return
	}

	data.ID = types.StringValue(wifiResp.ID)

	tflog.Debug(ctx, "Created UniFi WiFi broadcast", map[string]interface{}{
		"id": wifiResp.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WifiBroadcastResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WifiBroadcastResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading UniFi WiFi broadcast", map[string]interface{}{
		"site_id":           data.SiteID.ValueString(),
		"wifi_broadcast_id": data.ID.ValueString(),
	})

	wifiResp, err := r.client.GetWifiBroadcastDetails(ctx, networktypes.GetWifiBroadcastDetailsRequest{
		SiteID:          data.SiteID.ValueString(),
		WifiBroadcastID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read WiFi broadcast: %s", err))
		return
	}

	data.Name = types.StringValue(wifiResp.Name)
	data.Type = types.StringValue(wifiResp.Type)
	data.Enabled = types.BoolValue(wifiResp.Enabled)
	data.HideName = types.BoolValue(wifiResp.HideName)
	data.ClientIsolationEnabled = types.BoolValue(wifiResp.ClientIsolationEnabled)

	if wifiResp.Network != nil {
		data.NetworkID = types.StringValue(wifiResp.Network.NetworkID)
	}

	if wifiResp.SecurityConfiguration != nil {
		data.SecurityType = types.StringValue(wifiResp.SecurityConfiguration.Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WifiBroadcastResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WifiBroadcastResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating UniFi WiFi broadcast", map[string]interface{}{
		"site_id":           data.SiteID.ValueString(),
		"wifi_broadcast_id": data.ID.ValueString(),
	})

	updateReq := networktypes.UpdateWifiBroadcastRequest{
		SiteID:                 data.SiteID.ValueString(),
		WifiBroadcastID:        data.ID.ValueString(),
		Name:                   data.Name.ValueString(),
		Type:                   data.Type.ValueString(),
		Enabled:                data.Enabled.ValueBool(),
		HideName:               data.HideName.ValueBool(),
		ClientIsolationEnabled: data.ClientIsolationEnabled.ValueBool(),
	}

	if !data.NetworkID.IsNull() && !data.NetworkID.IsUnknown() {
		updateReq.Network = &networktypes.WifiNetworkReference{
			Type:      "network",
			NetworkID: data.NetworkID.ValueString(),
		}
	}

	if !data.SecurityType.IsNull() {
		securityConfig := &networktypes.WifiSecurityConfiguration{
			Type: data.SecurityType.ValueString(),
		}
		if !data.Passphrase.IsNull() && !data.Passphrase.IsUnknown() {
			securityConfig.Passphrase = data.Passphrase.ValueString()
		}
		updateReq.SecurityConfiguration = securityConfig
	}

	_, err := r.client.UpdateWifiBroadcast(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update WiFi broadcast: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WifiBroadcastResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WifiBroadcastResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting UniFi WiFi broadcast", map[string]interface{}{
		"site_id":           data.SiteID.ValueString(),
		"wifi_broadcast_id": data.ID.ValueString(),
	})

	err := r.client.DeleteWifiBroadcast(ctx, networktypes.DeleteWifiBroadcastRequest{
		SiteID:          data.SiteID.ValueString(),
		WifiBroadcastID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete WiFi broadcast: %s", err))
		return
	}
}

func (r *WifiBroadcastResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
