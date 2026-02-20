// Copyright (c) 2025 murasame29
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/murasame29/unifi-client-go/services/network"
	networktypes "github.com/murasame29/unifi-client-go/services/network/types"
)

var _ datasource.DataSource = &DeviceDataSource{}

func NewDeviceDataSource() datasource.DataSource {
	return &DeviceDataSource{}
}

type DeviceDataSource struct {
	client *network.Client
}

type DeviceDataSourceModel struct {
	SiteID          types.String `tfsdk:"site_id"`
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	MacAddress      types.String `tfsdk:"mac_address"`
	IPAddress       types.String `tfsdk:"ip_address"`
	Model           types.String `tfsdk:"model"`
	State           types.String `tfsdk:"state"`
	FirmwareVersion types.String `tfsdk:"firmware_version"`
	Supported       types.Bool   `tfsdk:"supported"`
}

func (d *DeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *DeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches details of a specific device.",
		Attributes: map[string]schema.Attribute{
			"site_id":          schema.StringAttribute{Required: true},
			"id":               schema.StringAttribute{Required: true},
			"name":             schema.StringAttribute{Computed: true},
			"mac_address":      schema.StringAttribute{Computed: true},
			"ip_address":       schema.StringAttribute{Computed: true},
			"model":            schema.StringAttribute{Computed: true},
			"state":            schema.StringAttribute{Computed: true},
			"firmware_version": schema.StringAttribute{Computed: true},
			"supported":        schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *DeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clients, ok := req.ProviderData.(*UnifiClients)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *UnifiClients, got: %T", req.ProviderData))
		return
	}
	d.client = clients.Network
}

func (d *DeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetAdoptedDeviceDetails(ctx, networktypes.GetAdoptedDeviceDetailsRequest{
		SiteID:   data.SiteID.ValueString(),
		DeviceID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device: %s", err))
		return
	}

	data.Name = types.StringValue(result.Name)
	data.MacAddress = types.StringValue(result.MacAddress)
	data.IPAddress = types.StringValue(result.IPAddress)
	data.Model = types.StringValue(result.Model)
	data.State = types.StringValue(result.State)
	data.FirmwareVersion = types.StringValue(result.FirmwareVersion)
	data.Supported = types.BoolValue(result.Supported)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
