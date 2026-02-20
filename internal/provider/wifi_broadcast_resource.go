// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	SiteID                              types.String `tfsdk:"site_id"`
	ID                                  types.String `tfsdk:"id"`
	Name                                types.String `tfsdk:"name"`
	Type                                types.String `tfsdk:"type"`
	Enabled                             types.Bool   `tfsdk:"enabled"`
	NetworkID                           types.String `tfsdk:"network_id"`
	SecurityConfiguration               types.Object `tfsdk:"security_configuration"`
	BroadcastingDeviceFilter            types.Object `tfsdk:"broadcasting_device_filter"`
	MulticastToUnicastConversionEnabled types.Bool   `tfsdk:"multicast_to_unicast_conversion_enabled"`
	ClientIsolationEnabled              types.Bool   `tfsdk:"client_isolation_enabled"`
	HideName                            types.Bool   `tfsdk:"hide_name"`
	UapsdEnabled                        types.Bool   `tfsdk:"uapsd_enabled"`
	BroadcastingFrequenciesGHz          types.List   `tfsdk:"broadcasting_frequencies_ghz"`
	MloEnabled                          types.Bool   `tfsdk:"mlo_enabled"`
	BandSteeringEnabled                 types.Bool   `tfsdk:"band_steering_enabled"`
	ArpProxyEnabled                     types.Bool   `tfsdk:"arp_proxy_enabled"`
	BssTransitionEnabled                types.Bool   `tfsdk:"bss_transition_enabled"`
	AdvertiseDeviceName                 types.Bool   `tfsdk:"advertise_device_name"`
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
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the WiFi broadcast.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
			"security_configuration": schema.SingleNestedAttribute{
				MarkdownDescription: "Security configuration for the WiFi broadcast.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Security type (open, wpa2, wpa3, wpa2wpa3).",
						Required:            true,
					},
					"passphrase": schema.StringAttribute{
						MarkdownDescription: "WiFi passphrase.",
						Optional:            true,
						Sensitive:           true,
					},
					"pmf_mode": schema.StringAttribute{
						MarkdownDescription: "Protected Management Frames mode (disabled, optional, required).",
						Optional:            true,
					},
					"fast_roaming_enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether fast roaming (802.11r) is enabled.",
						Optional:            true,
					},
					"group_rekey_interval_seconds": schema.Int64Attribute{
						MarkdownDescription: "Group rekey interval in seconds.",
						Optional:            true,
					},
					"radius_profile_id": schema.StringAttribute{
						MarkdownDescription: "RADIUS profile ID for enterprise authentication.",
						Optional:            true,
					},
					"coa_enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether RADIUS Change of Authorization is enabled.",
						Optional:            true,
					},
					"security_mode": schema.StringAttribute{
						MarkdownDescription: "Security mode.",
						Optional:            true,
					},
					"wpa3_fast_roaming_enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether WPA3 fast roaming is enabled.",
						Optional:            true,
					},
				},
			},
			"broadcasting_device_filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter for broadcasting devices.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Filter type (all, include, exclude).",
						Required:            true,
					},
					"device_ids": schema.ListAttribute{
						MarkdownDescription: "List of device IDs.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"device_tag_ids": schema.ListAttribute{
						MarkdownDescription: "List of device tag IDs.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"multicast_to_unicast_conversion_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether multicast to unicast conversion is enabled. Defaults to `false`.",
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
			"hide_name": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide the SSID. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"uapsd_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether U-APSD (Unscheduled Automatic Power Save Delivery) is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"broadcasting_frequencies_ghz": schema.ListAttribute{
				MarkdownDescription: "List of broadcasting frequencies in GHz (2.4, 5, 6).",
				Optional:            true,
				ElementType:         types.Float64Type,
			},
			"mlo_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Multi-Link Operation (WiFi 7) is enabled.",
				Optional:            true,
			},
			"band_steering_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether band steering is enabled.",
				Optional:            true,
			},
			"arp_proxy_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether ARP proxy is enabled.",
				Optional:            true,
			},
			"bss_transition_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether BSS transition (802.11v) is enabled.",
				Optional:            true,
			},
			"advertise_device_name": schema.BoolAttribute{
				MarkdownDescription: "Whether to advertise device name.",
				Optional:            true,
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
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *UnifiClients, got: %T", req.ProviderData))
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

	createReq := r.buildCreateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	wifiResp, err := r.client.CreateWifiBroadcast(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create WiFi broadcast: %s", err))
		return
	}

	data.ID = types.StringValue(wifiResp.ID)
	tflog.Debug(ctx, "Created UniFi WiFi broadcast", map[string]interface{}{"id": wifiResp.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WifiBroadcastResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WifiBroadcastResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifiResp, err := r.client.GetWifiBroadcastDetails(ctx, networktypes.GetWifiBroadcastDetailsRequest{
		SiteID:          data.SiteID.ValueString(),
		WifiBroadcastID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read WiFi broadcast: %s", err))
		return
	}

	r.mapResponseToModel(ctx, wifiResp, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WifiBroadcastResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WifiBroadcastResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := r.buildUpdateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

func (r *WifiBroadcastResource) buildCreateRequest(ctx context.Context, data *WifiBroadcastResourceModel, diags *diag.Diagnostics) networktypes.CreateWifiBroadcastRequest {
	createReq := networktypes.CreateWifiBroadcastRequest{
		SiteID:                              data.SiteID.ValueString(),
		Name:                                data.Name.ValueString(),
		Type:                                data.Type.ValueString(),
		Enabled:                             data.Enabled.ValueBool(),
		HideName:                            data.HideName.ValueBool(),
		ClientIsolationEnabled:              data.ClientIsolationEnabled.ValueBool(),
		MulticastToUnicastConversionEnabled: data.MulticastToUnicastConversionEnabled.ValueBool(),
		UapsdEnabled:                        data.UapsdEnabled.ValueBool(),
	}

	if !data.NetworkID.IsNull() && !data.NetworkID.IsUnknown() {
		createReq.Network = &networktypes.WifiNetworkReference{
			Type:      "network",
			NetworkID: data.NetworkID.ValueString(),
		}
	}

	if !data.SecurityConfiguration.IsNull() && !data.SecurityConfiguration.IsUnknown() {
		createReq.SecurityConfiguration = r.buildSecurityConfiguration(ctx, data.SecurityConfiguration, diags)
	}

	if !data.BroadcastingDeviceFilter.IsNull() && !data.BroadcastingDeviceFilter.IsUnknown() {
		createReq.BroadcastingDeviceFilter = r.buildBroadcastingDeviceFilter(ctx, data.BroadcastingDeviceFilter, diags)
	}

	if !data.BroadcastingFrequenciesGHz.IsNull() {
		var freqs []float64
		diags.Append(data.BroadcastingFrequenciesGHz.ElementsAs(ctx, &freqs, false)...)
		createReq.BroadcastingFrequenciesGHz = freqs
	}

	if !data.MloEnabled.IsNull() {
		mlo := data.MloEnabled.ValueBool()
		createReq.MloEnabled = &mlo
	}
	if !data.BandSteeringEnabled.IsNull() {
		bs := data.BandSteeringEnabled.ValueBool()
		createReq.BandSteeringEnabled = &bs
	}
	if !data.ArpProxyEnabled.IsNull() {
		arp := data.ArpProxyEnabled.ValueBool()
		createReq.ArpProxyEnabled = &arp
	}
	if !data.BssTransitionEnabled.IsNull() {
		bss := data.BssTransitionEnabled.ValueBool()
		createReq.BssTransitionEnabled = &bss
	}
	if !data.AdvertiseDeviceName.IsNull() {
		adv := data.AdvertiseDeviceName.ValueBool()
		createReq.AdvertiseDeviceName = &adv
	}

	return createReq
}

func (r *WifiBroadcastResource) buildUpdateRequest(ctx context.Context, data *WifiBroadcastResourceModel, diags *diag.Diagnostics) networktypes.UpdateWifiBroadcastRequest {
	updateReq := networktypes.UpdateWifiBroadcastRequest{
		SiteID:                              data.SiteID.ValueString(),
		WifiBroadcastID:                     data.ID.ValueString(),
		Name:                                data.Name.ValueString(),
		Type:                                data.Type.ValueString(),
		Enabled:                             data.Enabled.ValueBool(),
		HideName:                            data.HideName.ValueBool(),
		ClientIsolationEnabled:              data.ClientIsolationEnabled.ValueBool(),
		MulticastToUnicastConversionEnabled: data.MulticastToUnicastConversionEnabled.ValueBool(),
		UapsdEnabled:                        data.UapsdEnabled.ValueBool(),
	}

	if !data.NetworkID.IsNull() && !data.NetworkID.IsUnknown() {
		updateReq.Network = &networktypes.WifiNetworkReference{
			Type:      "network",
			NetworkID: data.NetworkID.ValueString(),
		}
	}

	if !data.SecurityConfiguration.IsNull() && !data.SecurityConfiguration.IsUnknown() {
		updateReq.SecurityConfiguration = r.buildSecurityConfiguration(ctx, data.SecurityConfiguration, diags)
	}

	if !data.BroadcastingDeviceFilter.IsNull() && !data.BroadcastingDeviceFilter.IsUnknown() {
		updateReq.BroadcastingDeviceFilter = r.buildBroadcastingDeviceFilter(ctx, data.BroadcastingDeviceFilter, diags)
	}

	if !data.BroadcastingFrequenciesGHz.IsNull() {
		var freqs []float64
		diags.Append(data.BroadcastingFrequenciesGHz.ElementsAs(ctx, &freqs, false)...)
		updateReq.BroadcastingFrequenciesGHz = freqs
	}

	if !data.MloEnabled.IsNull() {
		mlo := data.MloEnabled.ValueBool()
		updateReq.MloEnabled = &mlo
	}
	if !data.BandSteeringEnabled.IsNull() {
		bs := data.BandSteeringEnabled.ValueBool()
		updateReq.BandSteeringEnabled = &bs
	}
	if !data.ArpProxyEnabled.IsNull() {
		arp := data.ArpProxyEnabled.ValueBool()
		updateReq.ArpProxyEnabled = &arp
	}
	if !data.BssTransitionEnabled.IsNull() {
		bss := data.BssTransitionEnabled.ValueBool()
		updateReq.BssTransitionEnabled = &bss
	}
	if !data.AdvertiseDeviceName.IsNull() {
		adv := data.AdvertiseDeviceName.ValueBool()
		updateReq.AdvertiseDeviceName = &adv
	}

	return updateReq
}

type WifiSecurityConfigModel struct {
	Type                      types.String `tfsdk:"type"`
	Passphrase                types.String `tfsdk:"passphrase"`
	PmfMode                   types.String `tfsdk:"pmf_mode"`
	FastRoamingEnabled        types.Bool   `tfsdk:"fast_roaming_enabled"`
	GroupRekeyIntervalSeconds types.Int64  `tfsdk:"group_rekey_interval_seconds"`
	RadiusProfileID           types.String `tfsdk:"radius_profile_id"`
	CoaEnabled                types.Bool   `tfsdk:"coa_enabled"`
	SecurityMode              types.String `tfsdk:"security_mode"`
	Wpa3FastRoamingEnabled    types.Bool   `tfsdk:"wpa3_fast_roaming_enabled"`
}

func (r *WifiBroadcastResource) buildSecurityConfiguration(ctx context.Context, secObj types.Object, diags *diag.Diagnostics) *networktypes.WifiSecurityConfiguration {
	var secConfig WifiSecurityConfigModel
	diags.Append(secObj.As(ctx, &secConfig, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.WifiSecurityConfiguration{
		Type:         secConfig.Type.ValueString(),
		Passphrase:   secConfig.Passphrase.ValueString(),
		PmfMode:      secConfig.PmfMode.ValueString(),
		SecurityMode: secConfig.SecurityMode.ValueString(),
	}

	if !secConfig.FastRoamingEnabled.IsNull() {
		fr := secConfig.FastRoamingEnabled.ValueBool()
		result.FastRoamingEnabled = &fr
	}
	if !secConfig.GroupRekeyIntervalSeconds.IsNull() {
		gri := int(secConfig.GroupRekeyIntervalSeconds.ValueInt64())
		result.GroupRekeyIntervalSeconds = &gri
	}
	if !secConfig.RadiusProfileID.IsNull() && !secConfig.RadiusProfileID.IsUnknown() {
		result.RadiusConfiguration = &networktypes.WifiRadiusConfiguration{
			ProfileID: secConfig.RadiusProfileID.ValueString(),
		}
	}
	if !secConfig.CoaEnabled.IsNull() {
		coa := secConfig.CoaEnabled.ValueBool()
		result.CoaEnabled = &coa
	}
	if !secConfig.Wpa3FastRoamingEnabled.IsNull() {
		wpa3fr := secConfig.Wpa3FastRoamingEnabled.ValueBool()
		result.Wpa3FastRoamingEnabled = &wpa3fr
	}

	return result
}

type BroadcastingDeviceFilterModel struct {
	Type         types.String `tfsdk:"type"`
	DeviceIDs    types.List   `tfsdk:"device_ids"`
	DeviceTagIDs types.List   `tfsdk:"device_tag_ids"`
}

func (r *WifiBroadcastResource) buildBroadcastingDeviceFilter(ctx context.Context, filterObj types.Object, diags *diag.Diagnostics) *networktypes.BroadcastingDeviceFilter {
	var filter BroadcastingDeviceFilterModel
	diags.Append(filterObj.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.BroadcastingDeviceFilter{
		Type: filter.Type.ValueString(),
	}

	if !filter.DeviceIDs.IsNull() {
		var deviceIDs []string
		diags.Append(filter.DeviceIDs.ElementsAs(ctx, &deviceIDs, false)...)
		result.DeviceIDs = deviceIDs
	}
	if !filter.DeviceTagIDs.IsNull() {
		var tagIDs []string
		diags.Append(filter.DeviceTagIDs.ElementsAs(ctx, &tagIDs, false)...)
		result.DeviceTagIDs = tagIDs
	}

	return result
}

func (r *WifiBroadcastResource) mapResponseToModel(ctx context.Context, resp *networktypes.WifiBroadcast, data *WifiBroadcastResourceModel, diags *diag.Diagnostics) {
	data.Name = types.StringValue(resp.Name)
	data.Type = types.StringValue(resp.Type)
	data.Enabled = types.BoolValue(resp.Enabled)
	data.HideName = types.BoolValue(resp.HideName)
	data.ClientIsolationEnabled = types.BoolValue(resp.ClientIsolationEnabled)
	data.MulticastToUnicastConversionEnabled = types.BoolValue(resp.MulticastToUnicastConversionEnabled)
	data.UapsdEnabled = types.BoolValue(resp.UapsdEnabled)

	if resp.Network != nil {
		data.NetworkID = types.StringValue(resp.Network.NetworkID)
	}

	if resp.SecurityConfiguration != nil {
		secAttrTypes := map[string]attr.Type{
			"type":                         types.StringType,
			"passphrase":                   types.StringType,
			"pmf_mode":                     types.StringType,
			"fast_roaming_enabled":         types.BoolType,
			"group_rekey_interval_seconds": types.Int64Type,
			"radius_profile_id":            types.StringType,
			"coa_enabled":                  types.BoolType,
			"security_mode":                types.StringType,
			"wpa3_fast_roaming_enabled":    types.BoolType,
		}
		secAttrValues := map[string]attr.Value{
			"type":          types.StringValue(resp.SecurityConfiguration.Type),
			"passphrase":    types.StringValue(resp.SecurityConfiguration.Passphrase),
			"pmf_mode":      types.StringValue(resp.SecurityConfiguration.PmfMode),
			"security_mode": types.StringValue(resp.SecurityConfiguration.SecurityMode),
		}

		if resp.SecurityConfiguration.FastRoamingEnabled != nil {
			secAttrValues["fast_roaming_enabled"] = types.BoolValue(*resp.SecurityConfiguration.FastRoamingEnabled)
		} else {
			secAttrValues["fast_roaming_enabled"] = types.BoolNull()
		}
		if resp.SecurityConfiguration.GroupRekeyIntervalSeconds != nil {
			secAttrValues["group_rekey_interval_seconds"] = types.Int64Value(int64(*resp.SecurityConfiguration.GroupRekeyIntervalSeconds))
		} else {
			secAttrValues["group_rekey_interval_seconds"] = types.Int64Null()
		}
		if resp.SecurityConfiguration.RadiusConfiguration != nil {
			secAttrValues["radius_profile_id"] = types.StringValue(resp.SecurityConfiguration.RadiusConfiguration.ProfileID)
		} else {
			secAttrValues["radius_profile_id"] = types.StringNull()
		}
		if resp.SecurityConfiguration.CoaEnabled != nil {
			secAttrValues["coa_enabled"] = types.BoolValue(*resp.SecurityConfiguration.CoaEnabled)
		} else {
			secAttrValues["coa_enabled"] = types.BoolNull()
		}
		if resp.SecurityConfiguration.Wpa3FastRoamingEnabled != nil {
			secAttrValues["wpa3_fast_roaming_enabled"] = types.BoolValue(*resp.SecurityConfiguration.Wpa3FastRoamingEnabled)
		} else {
			secAttrValues["wpa3_fast_roaming_enabled"] = types.BoolNull()
		}

		secObj, d := types.ObjectValue(secAttrTypes, secAttrValues)
		diags.Append(d...)
		data.SecurityConfiguration = secObj
	}

	if resp.BroadcastingDeviceFilter != nil {
		filterAttrTypes := map[string]attr.Type{
			"type":           types.StringType,
			"device_ids":     types.ListType{ElemType: types.StringType},
			"device_tag_ids": types.ListType{ElemType: types.StringType},
		}
		filterAttrValues := map[string]attr.Value{
			"type": types.StringValue(resp.BroadcastingDeviceFilter.Type),
		}

		if len(resp.BroadcastingDeviceFilter.DeviceIDs) > 0 {
			deviceIDs, d := types.ListValueFrom(ctx, types.StringType, resp.BroadcastingDeviceFilter.DeviceIDs)
			diags.Append(d...)
			filterAttrValues["device_ids"] = deviceIDs
		} else {
			filterAttrValues["device_ids"] = types.ListNull(types.StringType)
		}

		if len(resp.BroadcastingDeviceFilter.DeviceTagIDs) > 0 {
			tagIDs, d := types.ListValueFrom(ctx, types.StringType, resp.BroadcastingDeviceFilter.DeviceTagIDs)
			diags.Append(d...)
			filterAttrValues["device_tag_ids"] = tagIDs
		} else {
			filterAttrValues["device_tag_ids"] = types.ListNull(types.StringType)
		}

		filterObj, d := types.ObjectValue(filterAttrTypes, filterAttrValues)
		diags.Append(d...)
		data.BroadcastingDeviceFilter = filterObj
	}

	if len(resp.BroadcastingFrequenciesGHz) > 0 {
		freqs, d := types.ListValueFrom(ctx, types.Float64Type, resp.BroadcastingFrequenciesGHz)
		diags.Append(d...)
		data.BroadcastingFrequenciesGHz = freqs
	}

	if resp.MloEnabled != nil {
		data.MloEnabled = types.BoolValue(*resp.MloEnabled)
	}
	if resp.BandSteeringEnabled != nil {
		data.BandSteeringEnabled = types.BoolValue(*resp.BandSteeringEnabled)
	}
	if resp.ArpProxyEnabled != nil {
		data.ArpProxyEnabled = types.BoolValue(*resp.ArpProxyEnabled)
	}
	if resp.BssTransitionEnabled != nil {
		data.BssTransitionEnabled = types.BoolValue(*resp.BssTransitionEnabled)
	}
	if resp.AdvertiseDeviceName != nil {
		data.AdvertiseDeviceName = types.BoolValue(*resp.AdvertiseDeviceName)
	}
}
