// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

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

var _ resource.Resource = &FirewallZoneResource{}
var _ resource.ResourceWithImportState = &FirewallZoneResource{}

func NewFirewallZoneResource() resource.Resource {
	return &FirewallZoneResource{}
}

type FirewallZoneResource struct {
	client *network.Client
}

type FirewallZoneResourceModel struct {
	SiteID     types.String `tfsdk:"site_id"`
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
}

func (r *FirewallZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone"
}

func (r *FirewallZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a UniFi firewall zone.",
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
				MarkdownDescription: "The name of the firewall zone.",
				Required:            true,
			},
			"network_ids": schema.ListAttribute{
				MarkdownDescription: "List of network IDs in this zone.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *FirewallZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FirewallZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating firewall zone", map[string]interface{}{"name": data.Name.ValueString()})

	var networkIDs []string
	if !data.NetworkIDs.IsNull() {
		resp.Diagnostics.Append(data.NetworkIDs.ElementsAs(ctx, &networkIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := networktypes.CreateFirewallZoneRequest{
		SiteID:     data.SiteID.ValueString(),
		Name:       data.Name.ValueString(),
		NetworkIDs: networkIDs,
	}

	result, err := r.client.CreateFirewallZone(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create firewall zone: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FirewallZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetFirewallZone(ctx, networktypes.GetFirewallZoneRequest{
		SiteID: data.SiteID.ValueString(),
		ZoneID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read firewall zone: %s", err))
		return
	}

	data.Name = types.StringValue(result.Name)
	networkIDs, diags := types.ListValueFrom(ctx, types.StringType, result.NetworkIDs)
	resp.Diagnostics.Append(diags...)
	data.NetworkIDs = networkIDs

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FirewallZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var networkIDs []string
	if !data.NetworkIDs.IsNull() {
		resp.Diagnostics.Append(data.NetworkIDs.ElementsAs(ctx, &networkIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateReq := networktypes.UpdateFirewallZoneRequest{
		SiteID:     data.SiteID.ValueString(),
		ZoneID:     data.ID.ValueString(),
		Name:       data.Name.ValueString(),
		NetworkIDs: networkIDs,
	}

	_, err := r.client.UpdateFirewallZone(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update firewall zone: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FirewallZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFirewallZone(ctx, networktypes.DeleteFirewallZoneRequest{
		SiteID: data.SiteID.ValueString(),
		ZoneID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete firewall zone: %s", err))
		return
	}
}

func (r *FirewallZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
