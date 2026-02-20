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

var _ resource.Resource = &ACLRuleResource{}
var _ resource.ResourceWithImportState = &ACLRuleResource{}

func NewACLRuleResource() resource.Resource {
	return &ACLRuleResource{}
}

type ACLRuleResource struct {
	client *network.Client
}

type ACLRuleResourceModel struct {
	SiteID      types.String `tfsdk:"site_id"`
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Action      types.String `tfsdk:"action"`
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
				MarkdownDescription: "The ACL rule type. Defaults to `wired`.",
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
				MarkdownDescription: "The action (allow/deny). Defaults to `allow`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("allow"),
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

	createReq := networktypes.CreateACLRuleRequest{
		SiteID:      data.SiteID.ValueString(),
		Type:        data.Type.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Action:      data.Action.ValueString(),
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

	data.Type = types.StringValue(result.Type)
	data.Name = types.StringValue(result.Name)
	data.Description = types.StringValue(result.Description)
	data.Enabled = types.BoolValue(result.Enabled)
	data.Action = types.StringValue(result.Action)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACLRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ACLRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := networktypes.UpdateACLRuleRequest{
		SiteID:      data.SiteID.ValueString(),
		RuleID:      data.ID.ValueString(),
		Type:        data.Type.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
		Action:      data.Action.ValueString(),
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
