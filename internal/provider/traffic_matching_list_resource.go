// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ resource.Resource = &TrafficMatchingListResource{}
var _ resource.ResourceWithImportState = &TrafficMatchingListResource{}

func NewTrafficMatchingListResource() resource.Resource {
	return &TrafficMatchingListResource{}
}

type TrafficMatchingListResource struct {
	client *network.Client
}

type TrafficMatchingListResourceModel struct {
	SiteID           types.String `tfsdk:"site_id"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	PortItems        types.List   `tfsdk:"port_items"`
	IPAddressItems   types.List   `tfsdk:"ip_address_items"`
	IPv6AddressItems types.List   `tfsdk:"ipv6_address_items"`
}

func (r *TrafficMatchingListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_matching_list"
}

func (r *TrafficMatchingListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi traffic matching list for use in firewall policies.",
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
				MarkdownDescription: "The name of the traffic matching list.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type (PORTS, IPV4_ADDRESSES, IPV6_ADDRESSES).",
				Required:            true,
			},
			"port_items": schema.ListNestedAttribute{
				MarkdownDescription: "Port items (for PORTS type).",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Item type (single, range).",
							Required:            true,
						},
						"value": schema.Int64Attribute{
							MarkdownDescription: "Single port value.",
							Optional:            true,
						},
						"start": schema.Int64Attribute{
							MarkdownDescription: "Range start port.",
							Optional:            true,
						},
						"stop": schema.Int64Attribute{
							MarkdownDescription: "Range stop port.",
							Optional:            true,
						},
					},
				},
			},
			"ip_address_items": schema.ListNestedAttribute{
				MarkdownDescription: "IPv4 address items (for IPV4_ADDRESSES type).",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Item type (single, range, subnet).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Single IP address or subnet.",
							Optional:            true,
						},
						"start": schema.StringAttribute{
							MarkdownDescription: "Range start IP address.",
							Optional:            true,
						},
						"stop": schema.StringAttribute{
							MarkdownDescription: "Range stop IP address.",
							Optional:            true,
						},
					},
				},
			},
			"ipv6_address_items": schema.ListNestedAttribute{
				MarkdownDescription: "IPv6 address items (for IPV6_ADDRESSES type).",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Item type (single, range, subnet).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Single IPv6 address or subnet.",
							Optional:            true,
						},
						"start": schema.StringAttribute{
							MarkdownDescription: "Range start IPv6 address.",
							Optional:            true,
						},
						"stop": schema.StringAttribute{
							MarkdownDescription: "Range stop IPv6 address.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *TrafficMatchingListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type PortItemModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.Int64  `tfsdk:"value"`
	Start types.Int64  `tfsdk:"start"`
	Stop  types.Int64  `tfsdk:"stop"`
}

type IPAddressItemModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
	Start types.String `tfsdk:"start"`
	Stop  types.String `tfsdk:"stop"`
}

func (r *TrafficMatchingListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TrafficMatchingListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating traffic matching list", map[string]interface{}{"name": data.Name.ValueString()})

	createReq := networktypes.CreateTrafficMatchingListRequest{
		SiteID: data.SiteID.ValueString(),
		Name:   data.Name.ValueString(),
		Type:   data.Type.ValueString(),
	}

	if !data.PortItems.IsNull() {
		var portItems []PortItemModel
		resp.Diagnostics.Append(data.PortItems.ElementsAs(ctx, &portItems, false)...)
		for _, item := range portItems {
			portItem := networktypes.PortMatchingItem{Type: item.Type.ValueString()}
			if !item.Value.IsNull() {
				v := int(item.Value.ValueInt64())
				portItem.Value = &v
			}
			if !item.Start.IsNull() {
				s := int(item.Start.ValueInt64())
				portItem.Start = &s
			}
			if !item.Stop.IsNull() {
				e := int(item.Stop.ValueInt64())
				portItem.Stop = &e
			}
			createReq.PortItems = append(createReq.PortItems, portItem)
		}
	}

	if !data.IPAddressItems.IsNull() {
		var ipItems []IPAddressItemModel
		resp.Diagnostics.Append(data.IPAddressItems.ElementsAs(ctx, &ipItems, false)...)
		for _, item := range ipItems {
			createReq.IPAddressItems = append(createReq.IPAddressItems, networktypes.IPAddressMatchingItem{
				Type:  item.Type.ValueString(),
				Value: item.Value.ValueString(),
				Start: item.Start.ValueString(),
				Stop:  item.Stop.ValueString(),
			})
		}
	}

	if !data.IPv6AddressItems.IsNull() {
		var ipv6Items []IPAddressItemModel
		resp.Diagnostics.Append(data.IPv6AddressItems.ElementsAs(ctx, &ipv6Items, false)...)
		for _, item := range ipv6Items {
			createReq.IPV6AddressItems = append(createReq.IPV6AddressItems, networktypes.IPAddressMatchingItem{
				Type:  item.Type.ValueString(),
				Value: item.Value.ValueString(),
				Start: item.Start.ValueString(),
				Stop:  item.Stop.ValueString(),
			})
		}
	}

	result, err := r.client.CreateTrafficMatchingList(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create traffic matching list: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TrafficMatchingListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TrafficMatchingListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetTrafficMatchingList(ctx, networktypes.GetTrafficMatchingListRequest{
		SiteID: data.SiteID.ValueString(),
		ListID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read traffic matching list: %s", err))
		return
	}

	data.Name = types.StringValue(result.Name)
	data.Type = types.StringValue(result.Type)

	portItemAttrTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.Int64Type,
		"start": types.Int64Type,
		"stop":  types.Int64Type,
	}

	if len(result.PortItems) > 0 {
		var portElements []attr.Value
		for _, item := range result.PortItems {
			attrValues := map[string]attr.Value{
				"type": types.StringValue(item.Type),
			}
			if item.Value != nil {
				attrValues["value"] = types.Int64Value(int64(*item.Value))
			} else {
				attrValues["value"] = types.Int64Null()
			}
			if item.Start != nil {
				attrValues["start"] = types.Int64Value(int64(*item.Start))
			} else {
				attrValues["start"] = types.Int64Null()
			}
			if item.Stop != nil {
				attrValues["stop"] = types.Int64Value(int64(*item.Stop))
			} else {
				attrValues["stop"] = types.Int64Null()
			}
			obj, d := types.ObjectValue(portItemAttrTypes, attrValues)
			resp.Diagnostics.Append(d...)
			portElements = append(portElements, obj)
		}
		portList, d := types.ListValue(types.ObjectType{AttrTypes: portItemAttrTypes}, portElements)
		resp.Diagnostics.Append(d...)
		data.PortItems = portList
	}

	ipItemAttrTypes := map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
		"start": types.StringType,
		"stop":  types.StringType,
	}

	if len(result.IPAddressItems) > 0 {
		var ipElements []attr.Value
		for _, item := range result.IPAddressItems {
			obj, d := types.ObjectValue(ipItemAttrTypes, map[string]attr.Value{
				"type":  types.StringValue(item.Type),
				"value": types.StringValue(item.Value),
				"start": types.StringValue(item.Start),
				"stop":  types.StringValue(item.Stop),
			})
			resp.Diagnostics.Append(d...)
			ipElements = append(ipElements, obj)
		}
		ipList, d := types.ListValue(types.ObjectType{AttrTypes: ipItemAttrTypes}, ipElements)
		resp.Diagnostics.Append(d...)
		data.IPAddressItems = ipList
	}

	if len(result.IPV6AddressItems) > 0 {
		var ipv6Elements []attr.Value
		for _, item := range result.IPV6AddressItems {
			obj, d := types.ObjectValue(ipItemAttrTypes, map[string]attr.Value{
				"type":  types.StringValue(item.Type),
				"value": types.StringValue(item.Value),
				"start": types.StringValue(item.Start),
				"stop":  types.StringValue(item.Stop),
			})
			resp.Diagnostics.Append(d...)
			ipv6Elements = append(ipv6Elements, obj)
		}
		ipv6List, d := types.ListValue(types.ObjectType{AttrTypes: ipItemAttrTypes}, ipv6Elements)
		resp.Diagnostics.Append(d...)
		data.IPv6AddressItems = ipv6List
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TrafficMatchingListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TrafficMatchingListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := networktypes.UpdateTrafficMatchingListRequest{
		SiteID: data.SiteID.ValueString(),
		ListID: data.ID.ValueString(),
		Name:   data.Name.ValueString(),
		Type:   data.Type.ValueString(),
	}

	if !data.PortItems.IsNull() {
		var portItems []PortItemModel
		resp.Diagnostics.Append(data.PortItems.ElementsAs(ctx, &portItems, false)...)
		for _, item := range portItems {
			portItem := networktypes.PortMatchingItem{Type: item.Type.ValueString()}
			if !item.Value.IsNull() {
				v := int(item.Value.ValueInt64())
				portItem.Value = &v
			}
			if !item.Start.IsNull() {
				s := int(item.Start.ValueInt64())
				portItem.Start = &s
			}
			if !item.Stop.IsNull() {
				e := int(item.Stop.ValueInt64())
				portItem.Stop = &e
			}
			updateReq.PortItems = append(updateReq.PortItems, portItem)
		}
	}

	if !data.IPAddressItems.IsNull() {
		var ipItems []IPAddressItemModel
		resp.Diagnostics.Append(data.IPAddressItems.ElementsAs(ctx, &ipItems, false)...)
		for _, item := range ipItems {
			updateReq.IPAddressItems = append(updateReq.IPAddressItems, networktypes.IPAddressMatchingItem{
				Type:  item.Type.ValueString(),
				Value: item.Value.ValueString(),
				Start: item.Start.ValueString(),
				Stop:  item.Stop.ValueString(),
			})
		}
	}

	if !data.IPv6AddressItems.IsNull() {
		var ipv6Items []IPAddressItemModel
		resp.Diagnostics.Append(data.IPv6AddressItems.ElementsAs(ctx, &ipv6Items, false)...)
		for _, item := range ipv6Items {
			updateReq.IPV6AddressItems = append(updateReq.IPV6AddressItems, networktypes.IPAddressMatchingItem{
				Type:  item.Type.ValueString(),
				Value: item.Value.ValueString(),
				Start: item.Start.ValueString(),
				Stop:  item.Stop.ValueString(),
			})
		}
	}

	_, err := r.client.UpdateTrafficMatchingList(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update traffic matching list: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TrafficMatchingListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TrafficMatchingListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTrafficMatchingList(ctx, networktypes.DeleteTrafficMatchingListRequest{
		SiteID: data.SiteID.ValueString(),
		ListID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete traffic matching list: %s", err))
		return
	}
}

func (r *TrafficMatchingListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
