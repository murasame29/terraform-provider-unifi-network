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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

type NetworkResource struct {
	client *network.Client
}

type NetworkResourceModel struct {
	SiteID                types.String `tfsdk:"site_id"`
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	VlanID                types.Int64  `tfsdk:"vlan_id"`
	Management            types.String `tfsdk:"management"`
	IsolationEnabled      types.Bool   `tfsdk:"isolation_enabled"`
	InternetAccessEnabled types.Bool   `tfsdk:"internet_access_enabled"`
	MdnsForwardingEnabled types.Bool   `tfsdk:"mdns_forwarding_enabled"`
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi network.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID where the network will be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the network.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the network.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the network is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"vlan_id": schema.Int64Attribute{
				MarkdownDescription: "The VLAN ID of the network. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"management": schema.StringAttribute{
				MarkdownDescription: "The management type of the network. Defaults to `third-party`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("third-party"),
			},
			"isolation_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether network isolation is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"internet_access_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether internet access is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"mdns_forwarding_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether mDNS forwarding is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating UniFi network", map[string]interface{}{
		"site_id": data.SiteID.ValueString(),
		"name":    data.Name.ValueString(),
	})

	isolationEnabled := data.IsolationEnabled.ValueBool()
	internetAccessEnabled := data.InternetAccessEnabled.ValueBool()
	mdnsForwardingEnabled := data.MdnsForwardingEnabled.ValueBool()

	createReq := networktypes.CreateNetworkRequest{
		SiteID:                data.SiteID.ValueString(),
		Name:                  data.Name.ValueString(),
		Enabled:               data.Enabled.ValueBool(),
		VlanID:                int(data.VlanID.ValueInt64()),
		Management:            data.Management.ValueString(),
		IsolationEnabled:      &isolationEnabled,
		InternetAccessEnabled: &internetAccessEnabled,
		MdnsForwardingEnabled: &mdnsForwardingEnabled,
	}

	networkResp, err := r.client.CreateNetwork(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network: %s", err))
		return
	}

	data.ID = types.StringValue(networkResp.ID)

	tflog.Debug(ctx, "Created UniFi network", map[string]interface{}{
		"id": networkResp.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading UniFi network", map[string]interface{}{
		"site_id":    data.SiteID.ValueString(),
		"network_id": data.ID.ValueString(),
	})

	networkResp, err := r.client.GetNetworkDetails(ctx, networktypes.GetNetworkDetailsRequest{
		SiteID:    data.SiteID.ValueString(),
		NetworkID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network: %s", err))
		return
	}

	data.Name = types.StringValue(networkResp.Name)
	data.Enabled = types.BoolValue(networkResp.Enabled)
	data.VlanID = types.Int64Value(int64(networkResp.VlanID))
	data.Management = types.StringValue(networkResp.Management)

	if networkResp.IsolationEnabled != nil {
		data.IsolationEnabled = types.BoolValue(*networkResp.IsolationEnabled)
	}
	if networkResp.InternetAccessEnabled != nil {
		data.InternetAccessEnabled = types.BoolValue(*networkResp.InternetAccessEnabled)
	}
	if networkResp.MdnsForwardingEnabled != nil {
		data.MdnsForwardingEnabled = types.BoolValue(*networkResp.MdnsForwardingEnabled)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating UniFi network", map[string]interface{}{
		"site_id":    data.SiteID.ValueString(),
		"network_id": data.ID.ValueString(),
	})

	isolationEnabled := data.IsolationEnabled.ValueBool()
	internetAccessEnabled := data.InternetAccessEnabled.ValueBool()
	mdnsForwardingEnabled := data.MdnsForwardingEnabled.ValueBool()

	updateReq := networktypes.UpdateNetworkRequest{
		SiteID:                data.SiteID.ValueString(),
		NetworkID:             data.ID.ValueString(),
		Name:                  data.Name.ValueString(),
		Enabled:               data.Enabled.ValueBool(),
		VlanID:                int(data.VlanID.ValueInt64()),
		Management:            data.Management.ValueString(),
		IsolationEnabled:      &isolationEnabled,
		InternetAccessEnabled: &internetAccessEnabled,
		MdnsForwardingEnabled: &mdnsForwardingEnabled,
	}

	_, err := r.client.UpdateNetwork(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update network: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting UniFi network", map[string]interface{}{
		"site_id":    data.SiteID.ValueString(),
		"network_id": data.ID.ValueString(),
	})

	err := r.client.DeleteNetwork(ctx, networktypes.DeleteNetworkRequest{
		SiteID:    data.SiteID.ValueString(),
		NetworkID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete network: %s", err))
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
