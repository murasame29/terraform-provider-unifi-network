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

var _ resource.Resource = &FirewallPolicyResource{}
var _ resource.ResourceWithImportState = &FirewallPolicyResource{}

func NewFirewallPolicyResource() resource.Resource {
	return &FirewallPolicyResource{}
}

type FirewallPolicyResource struct {
	client *network.Client
}

type FirewallPolicyResourceModel struct {
	SiteID                types.String `tfsdk:"site_id"`
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Action                types.Object `tfsdk:"action"`
	Source                types.Object `tfsdk:"source"`
	Destination           types.Object `tfsdk:"destination"`
	IPProtocolScope       types.Object `tfsdk:"ip_protocol_scope"`
	ConnectionStateFilter types.List   `tfsdk:"connection_state_filter"`
	IpsecFilter           types.String `tfsdk:"ipsec_filter"`
	LoggingEnabled        types.Bool   `tfsdk:"logging_enabled"`
	Schedule              types.Object `tfsdk:"schedule"`
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
			"action": schema.SingleNestedAttribute{
				MarkdownDescription: "The action configuration.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Action type (allow, drop, reject).",
						Required:            true,
					},
					"allow_return_traffic": schema.BoolAttribute{
						MarkdownDescription: "Whether to allow return traffic.",
						Optional:            true,
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: "Source endpoint configuration.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"zone_id": schema.StringAttribute{
						MarkdownDescription: "Source firewall zone ID.",
						Required:            true,
					},
					"traffic_filter": schema.SingleNestedAttribute{
						MarkdownDescription: "Traffic filter configuration.",
						Optional:            true,
						Attributes:          getTrafficFilterSchemaAttributes(),
					},
				},
			},
			"destination": schema.SingleNestedAttribute{
				MarkdownDescription: "Destination endpoint configuration.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"zone_id": schema.StringAttribute{
						MarkdownDescription: "Destination firewall zone ID.",
						Required:            true,
					},
					"traffic_filter": schema.SingleNestedAttribute{
						MarkdownDescription: "Traffic filter configuration.",
						Optional:            true,
						Attributes:          getTrafficFilterSchemaAttributes(),
					},
				},
			},
			"ip_protocol_scope": schema.SingleNestedAttribute{
				MarkdownDescription: "IP protocol scope configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"ip_version": schema.StringAttribute{
						MarkdownDescription: "IP version (ipv4, ipv6, both).",
						Required:            true,
					},
					"protocol_filter": schema.SingleNestedAttribute{
						MarkdownDescription: "Protocol filter configuration.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: "Filter type (protocol, protocol_number, preset).",
								Required:            true,
							},
							"protocol_name": schema.StringAttribute{
								MarkdownDescription: "Protocol name (tcp, udp, icmp, etc.).",
								Optional:            true,
							},
							"protocol_number": schema.Int64Attribute{
								MarkdownDescription: "Protocol number.",
								Optional:            true,
							},
							"preset_name": schema.StringAttribute{
								MarkdownDescription: "Preset name.",
								Optional:            true,
							},
							"match_opposite": schema.BoolAttribute{
								MarkdownDescription: "Whether to match opposite.",
								Optional:            true,
							},
						},
					},
				},
			},
			"connection_state_filter": schema.ListAttribute{
				MarkdownDescription: "Connection state filter (new, established, related, invalid).",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"ipsec_filter": schema.StringAttribute{
				MarkdownDescription: "IPsec filter (match-ipsec, match-none, any).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
			},
			"logging_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether logging is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "Schedule configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						MarkdownDescription: "Schedule mode (always, time-range).",
						Required:            true,
					},
					"repeat_on_days": schema.ListAttribute{
						MarkdownDescription: "Days to repeat (monday, tuesday, etc.).",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"start_date": schema.StringAttribute{
						MarkdownDescription: "Start date (YYYY-MM-DD).",
						Optional:            true,
					},
					"stop_date": schema.StringAttribute{
						MarkdownDescription: "Stop date (YYYY-MM-DD).",
						Optional:            true,
					},
					"start_time": schema.StringAttribute{
						MarkdownDescription: "Start time (HH:MM).",
						Optional:            true,
					},
					"stop_time": schema.StringAttribute{
						MarkdownDescription: "Stop time (HH:MM).",
						Optional:            true,
					},
				},
			},
		},
	}
}

func getTrafficFilterSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			MarkdownDescription: "Filter type.",
			Required:            true,
		},
		"port_filter": schema.SingleNestedAttribute{
			MarkdownDescription: "Port filter configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "Port filter type (items, traffic_matching_list).",
					Required:            true,
				},
				"match_opposite": schema.BoolAttribute{
					MarkdownDescription: "Whether to match opposite.",
					Optional:            true,
				},
				"traffic_matching_list_id": schema.StringAttribute{
					MarkdownDescription: "Traffic matching list ID.",
					Optional:            true,
				},
				"ports": schema.ListAttribute{
					MarkdownDescription: "List of ports.",
					Optional:            true,
					ElementType:         types.Int64Type,
				},
			},
		},
		"network_filter": schema.SingleNestedAttribute{
			MarkdownDescription: "Network filter configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"network_ids": schema.ListAttribute{
					MarkdownDescription: "List of network IDs.",
					Required:            true,
					ElementType:         types.StringType,
				},
				"match_opposite": schema.BoolAttribute{
					MarkdownDescription: "Whether to match opposite.",
					Optional:            true,
				},
			},
		},
		"ip_address_filter": schema.SingleNestedAttribute{
			MarkdownDescription: "IP address filter configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "IP address filter type (items, traffic_matching_list).",
					Required:            true,
				},
				"match_opposite": schema.BoolAttribute{
					MarkdownDescription: "Whether to match opposite.",
					Optional:            true,
				},
				"traffic_matching_list_id": schema.StringAttribute{
					MarkdownDescription: "Traffic matching list ID.",
					Optional:            true,
				},
				"addresses": schema.ListAttribute{
					MarkdownDescription: "List of IP addresses or subnets.",
					Optional:            true,
					ElementType:         types.StringType,
				},
			},
		},
		"region_filter": schema.SingleNestedAttribute{
			MarkdownDescription: "Region filter configuration.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"regions": schema.ListAttribute{
					MarkdownDescription: "List of region codes.",
					Required:            true,
					ElementType:         types.StringType,
				},
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

	createReq := r.buildCreateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

	r.mapResponseToModel(ctx, result, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := r.buildUpdateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

type FirewallActionModel struct {
	Type               types.String `tfsdk:"type"`
	AllowReturnTraffic types.Bool   `tfsdk:"allow_return_traffic"`
}

type FirewallEndpointModel struct {
	ZoneID        types.String `tfsdk:"zone_id"`
	TrafficFilter types.Object `tfsdk:"traffic_filter"`
}

type FirewallIPProtocolScopeModel struct {
	IPVersion      types.String `tfsdk:"ip_version"`
	ProtocolFilter types.Object `tfsdk:"protocol_filter"`
}

type FirewallProtocolFilterModel struct {
	Type           types.String `tfsdk:"type"`
	ProtocolName   types.String `tfsdk:"protocol_name"`
	ProtocolNumber types.Int64  `tfsdk:"protocol_number"`
	PresetName     types.String `tfsdk:"preset_name"`
	MatchOpposite  types.Bool   `tfsdk:"match_opposite"`
}

type FirewallScheduleModel struct {
	Mode         types.String `tfsdk:"mode"`
	RepeatOnDays types.List   `tfsdk:"repeat_on_days"`
	StartDate    types.String `tfsdk:"start_date"`
	StopDate     types.String `tfsdk:"stop_date"`
	StartTime    types.String `tfsdk:"start_time"`
	StopTime     types.String `tfsdk:"stop_time"`
}

func (r *FirewallPolicyResource) buildCreateRequest(ctx context.Context, data *FirewallPolicyResourceModel, diags *diag.Diagnostics) networktypes.CreateFirewallPolicyRequest {
	createReq := networktypes.CreateFirewallPolicyRequest{
		SiteID:         data.SiteID.ValueString(),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		LoggingEnabled: data.LoggingEnabled.ValueBool(),
		IpsecFilter:    data.IpsecFilter.ValueString(),
	}

	if !data.Action.IsNull() {
		createReq.Action = r.buildAction(ctx, data.Action, diags)
	}
	if !data.Source.IsNull() {
		createReq.Source = r.buildEndpoint(ctx, data.Source, diags)
	}
	if !data.Destination.IsNull() {
		createReq.Destination = r.buildEndpoint(ctx, data.Destination, diags)
	}
	if !data.IPProtocolScope.IsNull() {
		createReq.IPProtocolScope = r.buildIPProtocolScope(ctx, data.IPProtocolScope, diags)
	}
	if !data.ConnectionStateFilter.IsNull() {
		var states []string
		diags.Append(data.ConnectionStateFilter.ElementsAs(ctx, &states, false)...)
		createReq.ConnectionStateFilter = states
	}
	if !data.Schedule.IsNull() {
		createReq.Schedule = r.buildSchedule(ctx, data.Schedule, diags)
	}

	return createReq
}

func (r *FirewallPolicyResource) buildUpdateRequest(ctx context.Context, data *FirewallPolicyResourceModel, diags *diag.Diagnostics) networktypes.UpdateFirewallPolicyRequest {
	updateReq := networktypes.UpdateFirewallPolicyRequest{
		SiteID:         data.SiteID.ValueString(),
		PolicyID:       data.ID.ValueString(),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		LoggingEnabled: data.LoggingEnabled.ValueBool(),
		IpsecFilter:    data.IpsecFilter.ValueString(),
	}

	if !data.Action.IsNull() {
		updateReq.Action = r.buildAction(ctx, data.Action, diags)
	}
	if !data.Source.IsNull() {
		updateReq.Source = r.buildEndpoint(ctx, data.Source, diags)
	}
	if !data.Destination.IsNull() {
		updateReq.Destination = r.buildEndpoint(ctx, data.Destination, diags)
	}
	if !data.IPProtocolScope.IsNull() {
		updateReq.IPProtocolScope = r.buildIPProtocolScope(ctx, data.IPProtocolScope, diags)
	}
	if !data.ConnectionStateFilter.IsNull() {
		var states []string
		diags.Append(data.ConnectionStateFilter.ElementsAs(ctx, &states, false)...)
		updateReq.ConnectionStateFilter = states
	}
	if !data.Schedule.IsNull() {
		updateReq.Schedule = r.buildSchedule(ctx, data.Schedule, diags)
	}

	return updateReq
}

func (r *FirewallPolicyResource) buildAction(ctx context.Context, actionObj types.Object, diags *diag.Diagnostics) *networktypes.FirewallPolicyAction {
	var action FirewallActionModel
	diags.Append(actionObj.As(ctx, &action, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.FirewallPolicyAction{
		Type: action.Type.ValueString(),
	}
	if !action.AllowReturnTraffic.IsNull() {
		art := action.AllowReturnTraffic.ValueBool()
		result.AllowReturnTraffic = &art
	}
	return result
}

func (r *FirewallPolicyResource) buildEndpoint(ctx context.Context, endpointObj types.Object, diags *diag.Diagnostics) *networktypes.FirewallPolicyEndpoint {
	var endpoint FirewallEndpointModel
	diags.Append(endpointObj.As(ctx, &endpoint, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.FirewallPolicyEndpoint{
		ZoneID: endpoint.ZoneID.ValueString(),
	}
	// Traffic filter would be built here if needed
	return result
}

func (r *FirewallPolicyResource) buildIPProtocolScope(ctx context.Context, scopeObj types.Object, diags *diag.Diagnostics) *networktypes.FirewallIPProtocolScope {
	var scope FirewallIPProtocolScopeModel
	diags.Append(scopeObj.As(ctx, &scope, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.FirewallIPProtocolScope{
		IPVersion: scope.IPVersion.ValueString(),
	}

	if !scope.ProtocolFilter.IsNull() && !scope.ProtocolFilter.IsUnknown() {
		var pf FirewallProtocolFilterModel
		diags.Append(scope.ProtocolFilter.As(ctx, &pf, basetypes.ObjectAsOptions{})...)

		result.ProtocolFilter = &networktypes.FirewallProtocolFilter{
			Type: pf.Type.ValueString(),
		}
		if !pf.ProtocolName.IsNull() {
			result.ProtocolFilter.Protocol = &networktypes.FirewallProtocol{
				Name: pf.ProtocolName.ValueString(),
			}
		}
		if !pf.ProtocolNumber.IsNull() {
			pn := int(pf.ProtocolNumber.ValueInt64())
			result.ProtocolFilter.ProtocolNumber = &pn
		}
		if !pf.PresetName.IsNull() {
			result.ProtocolFilter.Preset = &networktypes.FirewallProtocolPreset{
				Name: pf.PresetName.ValueString(),
			}
		}
		if !pf.MatchOpposite.IsNull() {
			mo := pf.MatchOpposite.ValueBool()
			result.ProtocolFilter.MatchOpposite = &mo
		}
	}

	return result
}

func (r *FirewallPolicyResource) buildSchedule(ctx context.Context, scheduleObj types.Object, diags *diag.Diagnostics) *networktypes.FirewallSchedule {
	var schedule FirewallScheduleModel
	diags.Append(scheduleObj.As(ctx, &schedule, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.FirewallSchedule{
		Mode:      schedule.Mode.ValueString(),
		StartDate: schedule.StartDate.ValueString(),
		StopDate:  schedule.StopDate.ValueString(),
	}

	if !schedule.RepeatOnDays.IsNull() {
		var days []string
		diags.Append(schedule.RepeatOnDays.ElementsAs(ctx, &days, false)...)
		result.RepeatOnDays = days
	}

	if !schedule.StartTime.IsNull() || !schedule.StopTime.IsNull() {
		result.TimeFilter = &networktypes.FirewallTimeFilter{
			StartTime: schedule.StartTime.ValueString(),
			StopTime:  schedule.StopTime.ValueString(),
		}
	}

	return result
}

func (r *FirewallPolicyResource) mapResponseToModel(ctx context.Context, resp *networktypes.FirewallPolicy, data *FirewallPolicyResourceModel, diags *diag.Diagnostics) {
	data.Name = types.StringValue(resp.Name)
	data.Description = types.StringValue(resp.Description)
	data.Enabled = types.BoolValue(resp.Enabled)
	data.LoggingEnabled = types.BoolValue(resp.LoggingEnabled)
	data.IpsecFilter = types.StringValue(resp.IpsecFilter)

	if resp.Action != nil {
		actionAttrTypes := map[string]attr.Type{
			"type":                 types.StringType,
			"allow_return_traffic": types.BoolType,
		}
		actionAttrValues := map[string]attr.Value{
			"type": types.StringValue(resp.Action.Type),
		}
		if resp.Action.AllowReturnTraffic != nil {
			actionAttrValues["allow_return_traffic"] = types.BoolValue(*resp.Action.AllowReturnTraffic)
		} else {
			actionAttrValues["allow_return_traffic"] = types.BoolNull()
		}
		actionObj, d := types.ObjectValue(actionAttrTypes, actionAttrValues)
		diags.Append(d...)
		data.Action = actionObj
	}

	if resp.Source != nil {
		data.Source = r.mapEndpointToObject(ctx, resp.Source, diags)
	}
	if resp.Destination != nil {
		data.Destination = r.mapEndpointToObject(ctx, resp.Destination, diags)
	}
	if resp.IPProtocolScope != nil {
		data.IPProtocolScope = r.mapIPProtocolScopeToObject(ctx, resp.IPProtocolScope, diags)
	}
	if len(resp.ConnectionStateFilter) > 0 {
		states, d := types.ListValueFrom(ctx, types.StringType, resp.ConnectionStateFilter)
		diags.Append(d...)
		data.ConnectionStateFilter = states
	}
	if resp.Schedule != nil {
		data.Schedule = r.mapScheduleToObject(ctx, resp.Schedule, diags)
	}
}

func (r *FirewallPolicyResource) mapEndpointToObject(ctx context.Context, endpoint *networktypes.FirewallPolicyEndpoint, diags *diag.Diagnostics) types.Object {
	attrTypes := map[string]attr.Type{
		"zone_id":        types.StringType,
		"traffic_filter": types.ObjectType{AttrTypes: getTrafficFilterAttrTypes()},
	}
	attrValues := map[string]attr.Value{
		"zone_id":        types.StringValue(endpoint.ZoneID),
		"traffic_filter": types.ObjectNull(getTrafficFilterAttrTypes()),
	}

	obj, d := types.ObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	return obj
}

func getTrafficFilterAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"port_filter": types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":                     types.StringType,
			"match_opposite":           types.BoolType,
			"traffic_matching_list_id": types.StringType,
			"ports":                    types.ListType{ElemType: types.Int64Type},
		}},
		"network_filter": types.ObjectType{AttrTypes: map[string]attr.Type{
			"network_ids":    types.ListType{ElemType: types.StringType},
			"match_opposite": types.BoolType,
		}},
		"ip_address_filter": types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":                     types.StringType,
			"match_opposite":           types.BoolType,
			"traffic_matching_list_id": types.StringType,
			"addresses":                types.ListType{ElemType: types.StringType},
		}},
		"region_filter": types.ObjectType{AttrTypes: map[string]attr.Type{
			"regions": types.ListType{ElemType: types.StringType},
		}},
	}
}

func (r *FirewallPolicyResource) mapIPProtocolScopeToObject(ctx context.Context, scope *networktypes.FirewallIPProtocolScope, diags *diag.Diagnostics) types.Object {
	protocolFilterAttrTypes := map[string]attr.Type{
		"type":            types.StringType,
		"protocol_name":   types.StringType,
		"protocol_number": types.Int64Type,
		"preset_name":     types.StringType,
		"match_opposite":  types.BoolType,
	}

	attrTypes := map[string]attr.Type{
		"ip_version":      types.StringType,
		"protocol_filter": types.ObjectType{AttrTypes: protocolFilterAttrTypes},
	}

	attrValues := map[string]attr.Value{
		"ip_version": types.StringValue(scope.IPVersion),
	}

	if scope.ProtocolFilter != nil {
		pfAttrValues := map[string]attr.Value{
			"type": types.StringValue(scope.ProtocolFilter.Type),
		}
		if scope.ProtocolFilter.Protocol != nil {
			pfAttrValues["protocol_name"] = types.StringValue(scope.ProtocolFilter.Protocol.Name)
		} else {
			pfAttrValues["protocol_name"] = types.StringNull()
		}
		if scope.ProtocolFilter.ProtocolNumber != nil {
			pfAttrValues["protocol_number"] = types.Int64Value(int64(*scope.ProtocolFilter.ProtocolNumber))
		} else {
			pfAttrValues["protocol_number"] = types.Int64Null()
		}
		if scope.ProtocolFilter.Preset != nil {
			pfAttrValues["preset_name"] = types.StringValue(scope.ProtocolFilter.Preset.Name)
		} else {
			pfAttrValues["preset_name"] = types.StringNull()
		}
		if scope.ProtocolFilter.MatchOpposite != nil {
			pfAttrValues["match_opposite"] = types.BoolValue(*scope.ProtocolFilter.MatchOpposite)
		} else {
			pfAttrValues["match_opposite"] = types.BoolNull()
		}

		pfObj, d := types.ObjectValue(protocolFilterAttrTypes, pfAttrValues)
		diags.Append(d...)
		attrValues["protocol_filter"] = pfObj
	} else {
		attrValues["protocol_filter"] = types.ObjectNull(protocolFilterAttrTypes)
	}

	obj, d := types.ObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	return obj
}

func (r *FirewallPolicyResource) mapScheduleToObject(ctx context.Context, schedule *networktypes.FirewallSchedule, diags *diag.Diagnostics) types.Object {
	attrTypes := map[string]attr.Type{
		"mode":           types.StringType,
		"repeat_on_days": types.ListType{ElemType: types.StringType},
		"start_date":     types.StringType,
		"stop_date":      types.StringType,
		"start_time":     types.StringType,
		"stop_time":      types.StringType,
	}

	attrValues := map[string]attr.Value{
		"mode":       types.StringValue(schedule.Mode),
		"start_date": types.StringValue(schedule.StartDate),
		"stop_date":  types.StringValue(schedule.StopDate),
	}

	if len(schedule.RepeatOnDays) > 0 {
		days, d := types.ListValueFrom(ctx, types.StringType, schedule.RepeatOnDays)
		diags.Append(d...)
		attrValues["repeat_on_days"] = days
	} else {
		attrValues["repeat_on_days"] = types.ListNull(types.StringType)
	}

	if schedule.TimeFilter != nil {
		attrValues["start_time"] = types.StringValue(schedule.TimeFilter.StartTime)
		attrValues["stop_time"] = types.StringValue(schedule.TimeFilter.StopTime)
	} else {
		attrValues["start_time"] = types.StringNull()
		attrValues["stop_time"] = types.StringNull()
	}

	obj, d := types.ObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	return obj
}
