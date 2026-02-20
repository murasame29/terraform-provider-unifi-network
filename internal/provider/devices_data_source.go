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

var _ datasource.DataSource = &DevicesDataSource{}

func NewDevicesDataSource() datasource.DataSource {
	return &DevicesDataSource{}
}

type DevicesDataSource struct {
	client *network.Client
}

type DevicesDataSourceModel struct {
	SiteID  types.String        `tfsdk:"site_id"`
	Devices []DeviceSummaryModel `tfsdk:"devices"`
}

type DeviceSummaryModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	MacAddress      types.String `tfsdk:"mac_address"`
	IPAddress       types.String `tfsdk:"ip_address"`
	Model           types.String `tfsdk:"model"`
	State           types.String `tfsdk:"state"`
	FirmwareVersion types.String `tfsdk:"firmware_version"`
}

func (d *DevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *DevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of adopted devices for a site.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "The site ID.",
				Required:            true,
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "List of devices.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"name":             schema.StringAttribute{Computed: true},
						"mac_address":      schema.StringAttribute{Computed: true},
						"ip_address":       schema.StringAttribute{Computed: true},
						"model":            schema.StringAttribute{Computed: true},
						"state":            schema.StringAttribute{Computed: true},
						"firmware_version": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *DevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListAdoptedDevices(ctx, networktypes.ListAdoptedDevicesRequest{
		SiteID: data.SiteID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read devices: %s", err))
		return
	}

	data.Devices = make([]DeviceSummaryModel, 0, len(result.Data))
	for _, device := range result.Data {
		data.Devices = append(data.Devices, DeviceSummaryModel{
			ID:              types.StringValue(device.ID),
			Name:            types.StringValue(device.Name),
			MacAddress:      types.StringValue(device.MacAddress),
			IPAddress:       types.StringValue(device.IPAddress),
			Model:           types.StringValue(device.Model),
			State:           types.StringValue(device.State),
			FirmwareVersion: types.StringValue(device.FirmwareVersion),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
