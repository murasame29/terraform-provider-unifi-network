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

var _ resource.Resource = &ACLRuleResource{}
var _ resource.ResourceWithImportState = &ACLRuleResource{}

func NewACLRuleResource() resource.Resource {
	return &ACLRuleResource{}
}

type ACLRuleResource struct {
	client *network.Client
}

type ACLRuleResourceModel struct {
	SiteID                types.String `tfsdk:"site_id"`
	ID                    types.String `tfsdk:"id"`
	Type                  types.String `tfsdk:"type"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Action                types.String `tfsdk:"action"`
	Index                 types.Int64  `tfsdk:"index"`
	EnforcingDeviceFilter types.Object `tfsdk:"enforcing_device_filter"`
	SourceFilter          types.Object `tfsdk:"source_filter"`
	DestinationFilter     types.Object `tfsdk:"destination_filter"`
	ProtocolFilter        types.List   `tfsdk:"protocol_filter"`
	NetworkIDFilter       types.String `tfsdk:"network_id_filter"`
}

func (r *ACLRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_rule"
}

func (r *ACLRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi ACL rule.",
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
				MarkdownDescription: "The ACL rule type (wired, wireless). Defaults to `wired`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("wired"),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the ACL rule.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The action (allow, deny). Defaults to `allow`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("allow"),
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "The rule index (order).",
				Optional:            true,
			},
			"enforcing_device_filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter for enforcing devices.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Filter type (all, include).",
						Required:            true,
					},
					"device_ids": schema.ListAttribute{
						MarkdownDescription: "List of device IDs.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"source_filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Source endpoint filter.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Filter type (any, ip_addresses, networks, mac_addresses).",
						Required:            true,
					},
					"ip_addresses_or_subnets": schema.ListAttribute{
						MarkdownDescription: "List of IP addresses or subnets.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"network_ids": schema.ListAttribute{
						MarkdownDescription: "List of network IDs.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"mac_addresses": schema.ListAttribute{
						MarkdownDescription: "List of MAC addresses.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"port_filter": schema.ListAttribute{
						MarkdownDescription: "List of ports.",
						Optional:            true,
						ElementType:         types.Int64Type,
					},
					"prefix_length": schema.Int64Attribute{
						MarkdownDescription: "Prefix length for IPv6.",
						Optional:            true,
					},
				},
			},
			"destination_filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Destination endpoint filter.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Filter type (any, ip_addresses, networks, mac_addresses).",
						Required:            true,
					},
					"ip_addresses_or_subnets": schema.ListAttribute{
						MarkdownDescription: "List of IP addresses or subnets.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"network_ids": schema.ListAttribute{
						MarkdownDescription: "List of network IDs.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"mac_addresses": schema.ListAttribute{
						MarkdownDescription: "List of MAC addresses.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"port_filter": schema.ListAttribute{
						MarkdownDescription: "List of ports.",
						Optional:            true,
						ElementType:         types.Int64Type,
					},
					"prefix_length": schema.Int64Attribute{
						MarkdownDescription: "Prefix length for IPv6.",
						Optional:            true,
					},
				},
			},
			"protocol_filter": schema.ListAttribute{
				MarkdownDescription: "List of protocols (tcp, udp, icmp, etc.).",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"network_id_filter": schema.StringAttribute{
				MarkdownDescription: "Network ID filter.",
				Optional:            true,
			},
		},
	}
}

func (r *ACLRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACLRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ACLRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating ACL rule", map[string]interface{}{"name": data.Name.ValueString()})

	createReq := r.buildCreateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateACLRule(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ACL rule: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACLRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ACLRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetACLRule(ctx, networktypes.GetACLRuleRequest{
		SiteID: data.SiteID.ValueString(),
		RuleID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ACL rule: %s", err))
		return
	}

	r.mapResponseToModel(ctx, result, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACLRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ACLRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := r.buildUpdateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateACLRule(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update ACL rule: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACLRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ACLRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteACLRule(ctx, networktypes.DeleteACLRuleRequest{
		SiteID: data.SiteID.ValueString(),
		RuleID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete ACL rule: %s", err))
		return
	}
}

func (r *ACLRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type ACLDeviceFilterModel struct {
	Type      types.String `tfsdk:"type"`
	DeviceIDs types.List   `tfsdk:"device_ids"`
}

type ACLEndpointFilterModel struct {
	Type                 types.String `tfsdk:"type"`
	IpAddressesOrSubnets types.List   `tfsdk:"ip_addresses_or_subnets"`
	NetworkIDs           types.List   `tfsdk:"network_ids"`
	MacAddresses         types.List   `tfsdk:"mac_addresses"`
	PortFilter           types.List   `tfsdk:"port_filter"`
	PrefixLength         types.Int64  `tfsdk:"prefix_length"`
}

func (r *ACLRuleResource) buildCreateRequest(ctx context.Context, data *ACLRuleResourceModel, diags *diag.Diagnostics) networktypes.CreateACLRuleRequest {
	createReq := networktypes.CreateACLRuleRequest{
		SiteID:          data.SiteID.ValueString(),
		Type:            data.Type.ValueString(),
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		Enabled:         data.Enabled.ValueBool(),
		Action:          data.Action.ValueString(),
		Index:           int(data.Index.ValueInt64()),
		NetworkIdFilter: data.NetworkIDFilter.ValueString(),
	}

	if !data.EnforcingDeviceFilter.IsNull() {
		createReq.EnforcingDeviceFilter = r.buildDeviceFilter(ctx, data.EnforcingDeviceFilter, diags)
	}
	if !data.SourceFilter.IsNull() {
		createReq.SourceFilter = r.buildEndpointFilter(ctx, data.SourceFilter, diags)
	}
	if !data.DestinationFilter.IsNull() {
		createReq.DestinationFilter = r.buildEndpointFilter(ctx, data.DestinationFilter, diags)
	}
	if !data.ProtocolFilter.IsNull() {
		var protocols []string
		diags.Append(data.ProtocolFilter.ElementsAs(ctx, &protocols, false)...)
		createReq.ProtocolFilter = protocols
	}

	return createReq
}

func (r *ACLRuleResource) buildUpdateRequest(ctx context.Context, data *ACLRuleResourceModel, diags *diag.Diagnostics) networktypes.UpdateACLRuleRequest {
	updateReq := networktypes.UpdateACLRuleRequest{
		SiteID:          data.SiteID.ValueString(),
		RuleID:          data.ID.ValueString(),
		Type:            data.Type.ValueString(),
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		Enabled:         data.Enabled.ValueBool(),
		Action:          data.Action.ValueString(),
		Index:           int(data.Index.ValueInt64()),
		NetworkIdFilter: data.NetworkIDFilter.ValueString(),
	}

	if !data.EnforcingDeviceFilter.IsNull() {
		updateReq.EnforcingDeviceFilter = r.buildDeviceFilter(ctx, data.EnforcingDeviceFilter, diags)
	}
	if !data.SourceFilter.IsNull() {
		updateReq.SourceFilter = r.buildEndpointFilter(ctx, data.SourceFilter, diags)
	}
	if !data.DestinationFilter.IsNull() {
		updateReq.DestinationFilter = r.buildEndpointFilter(ctx, data.DestinationFilter, diags)
	}
	if !data.ProtocolFilter.IsNull() {
		var protocols []string
		diags.Append(data.ProtocolFilter.ElementsAs(ctx, &protocols, false)...)
		updateReq.ProtocolFilter = protocols
	}

	return updateReq
}

func (r *ACLRuleResource) buildDeviceFilter(ctx context.Context, filterObj types.Object, diags *diag.Diagnostics) *networktypes.ACLDeviceFilter {
	var filter ACLDeviceFilterModel
	diags.Append(filterObj.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.ACLDeviceFilter{
		Type: filter.Type.ValueString(),
	}
	if !filter.DeviceIDs.IsNull() {
		var deviceIDs []string
		diags.Append(filter.DeviceIDs.ElementsAs(ctx, &deviceIDs, false)...)
		result.DeviceIDs = deviceIDs
	}
	return result
}

func (r *ACLRuleResource) buildEndpointFilter(ctx context.Context, filterObj types.Object, diags *diag.Diagnostics) *networktypes.ACLEndpointFilter {
	var filter ACLEndpointFilterModel
	diags.Append(filterObj.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.ACLEndpointFilter{
		Type: filter.Type.ValueString(),
	}
	if !filter.IpAddressesOrSubnets.IsNull() {
		var ips []string
		diags.Append(filter.IpAddressesOrSubnets.ElementsAs(ctx, &ips, false)...)
		result.IpAddressesOrSubnets = ips
	}
	if !filter.NetworkIDs.IsNull() {
		var networkIDs []string
		diags.Append(filter.NetworkIDs.ElementsAs(ctx, &networkIDs, false)...)
		result.NetworkIDs = networkIDs
	}
	if !filter.MacAddresses.IsNull() {
		var macs []string
		diags.Append(filter.MacAddresses.ElementsAs(ctx, &macs, false)...)
		result.MacAddresses = macs
	}
	if !filter.PortFilter.IsNull() {
		var ports []int64
		diags.Append(filter.PortFilter.ElementsAs(ctx, &ports, false)...)
		for _, p := range ports {
			result.PortFilter = append(result.PortFilter, int(p))
		}
	}
	if !filter.PrefixLength.IsNull() {
		pl := int(filter.PrefixLength.ValueInt64())
		result.PrefixLength = &pl
	}
	return result
}

func (r *ACLRuleResource) mapResponseToModel(ctx context.Context, resp *networktypes.ACLRule, data *ACLRuleResourceModel, diags *diag.Diagnostics) {
	data.Type = types.StringValue(resp.Type)
	data.Name = types.StringValue(resp.Name)
	data.Description = types.StringValue(resp.Description)
	data.Enabled = types.BoolValue(resp.Enabled)
	data.Action = types.StringValue(resp.Action)
	data.Index = types.Int64Value(int64(resp.Index))
	data.NetworkIDFilter = types.StringValue(resp.NetworkIdFilter)

	if resp.EnforcingDeviceFilter != nil {
		deviceFilterAttrTypes := map[string]attr.Type{
			"type":       types.StringType,
			"device_ids": types.ListType{ElemType: types.StringType},
		}
		deviceFilterAttrValues := map[string]attr.Value{
			"type": types.StringValue(resp.EnforcingDeviceFilter.Type),
		}
		if len(resp.EnforcingDeviceFilter.DeviceIDs) > 0 {
			deviceIDs, d := types.ListValueFrom(ctx, types.StringType, resp.EnforcingDeviceFilter.DeviceIDs)
			diags.Append(d...)
			deviceFilterAttrValues["device_ids"] = deviceIDs
		} else {
			deviceFilterAttrValues["device_ids"] = types.ListNull(types.StringType)
		}
		filterObj, d := types.ObjectValue(deviceFilterAttrTypes, deviceFilterAttrValues)
		diags.Append(d...)
		data.EnforcingDeviceFilter = filterObj
	}

	if resp.SourceFilter != nil {
		data.SourceFilter = r.mapEndpointFilterToObject(ctx, resp.SourceFilter, diags)
	}
	if resp.DestinationFilter != nil {
		data.DestinationFilter = r.mapEndpointFilterToObject(ctx, resp.DestinationFilter, diags)
	}
	if len(resp.ProtocolFilter) > 0 {
		protocols, d := types.ListValueFrom(ctx, types.StringType, resp.ProtocolFilter)
		diags.Append(d...)
		data.ProtocolFilter = protocols
	}
}

func (r *ACLRuleResource) mapEndpointFilterToObject(ctx context.Context, filter *networktypes.ACLEndpointFilter, diags *diag.Diagnostics) types.Object {
	attrTypes := map[string]attr.Type{
		"type":                    types.StringType,
		"ip_addresses_or_subnets": types.ListType{ElemType: types.StringType},
		"network_ids":             types.ListType{ElemType: types.StringType},
		"mac_addresses":           types.ListType{ElemType: types.StringType},
		"port_filter":             types.ListType{ElemType: types.Int64Type},
		"prefix_length":           types.Int64Type,
	}
	attrValues := map[string]attr.Value{
		"type": types.StringValue(filter.Type),
	}

	if len(filter.IpAddressesOrSubnets) > 0 {
		ips, d := types.ListValueFrom(ctx, types.StringType, filter.IpAddressesOrSubnets)
		diags.Append(d...)
		attrValues["ip_addresses_or_subnets"] = ips
	} else {
		attrValues["ip_addresses_or_subnets"] = types.ListNull(types.StringType)
	}
	if len(filter.NetworkIDs) > 0 {
		networkIDs, d := types.ListValueFrom(ctx, types.StringType, filter.NetworkIDs)
		diags.Append(d...)
		attrValues["network_ids"] = networkIDs
	} else {
		attrValues["network_ids"] = types.ListNull(types.StringType)
	}
	if len(filter.MacAddresses) > 0 {
		macs, d := types.ListValueFrom(ctx, types.StringType, filter.MacAddresses)
		diags.Append(d...)
		attrValues["mac_addresses"] = macs
	} else {
		attrValues["mac_addresses"] = types.ListNull(types.StringType)
	}
	if len(filter.PortFilter) > 0 {
		var ports []int64
		for _, p := range filter.PortFilter {
			ports = append(ports, int64(p))
		}
		portList, d := types.ListValueFrom(ctx, types.Int64Type, ports)
		diags.Append(d...)
		attrValues["port_filter"] = portList
	} else {
		attrValues["port_filter"] = types.ListNull(types.Int64Type)
	}
	if filter.PrefixLength != nil {
		attrValues["prefix_length"] = types.Int64Value(int64(*filter.PrefixLength))
	} else {
		attrValues["prefix_length"] = types.Int64Null()
	}

	obj, d := types.ObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	return obj
}
