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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

type NetworkDHCPIPAddressRangeModel struct {
	Start types.String `tfsdk:"start"`
	Stop  types.String `tfsdk:"stop"`
}

type NetworkPXEConfigurationModel struct {
	ServerIPAddress types.String `tfsdk:"server_ip_address"`
	Filename        types.String `tfsdk:"filename"`
}

type NetworkDHCPConfigurationModel struct {
	Mode                         types.String `tfsdk:"mode"`
	IPAddressRange               types.Object `tfsdk:"ip_address_range"`
	GatewayIPAddressOverride     types.String `tfsdk:"gateway_ip_address_override"`
	DNSServerIPAddressesOverride types.List   `tfsdk:"dns_server_ip_addresses_override"`
	LeaseTimeSeconds             types.Int64  `tfsdk:"lease_time_seconds"`
	DomainName                   types.String `tfsdk:"domain_name"`
	PingConflictDetectionEnabled types.Bool   `tfsdk:"ping_conflict_detection_enabled"`
	PxeConfiguration             types.Object `tfsdk:"pxe_configuration"`
	NtpServerIPAddresses         types.List   `tfsdk:"ntp_server_ip_addresses"`
	Option43Value                types.String `tfsdk:"option43_value"`
	TftpServerAddress            types.String `tfsdk:"tftp_server_address"`
	TimeOffsetSeconds            types.Int64  `tfsdk:"time_offset_seconds"`
	WpadURL                      types.String `tfsdk:"wpad_url"`
	WinsServerIPAddresses        types.List   `tfsdk:"wins_server_ip_addresses"`
	DHCPServerIPAddresses        types.List   `tfsdk:"dhcp_server_ip_addresses"`
}

type IPAddressSelectorModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type NetworkNATOutboundIPAddressConfigModel struct {
	Type               types.String `tfsdk:"type"`
	WanInterfaceID     types.String `tfsdk:"wan_interface_id"`
	IpAddressSelectors types.List   `tfsdk:"ip_address_selectors"`
}

type NetworkIPv4ConfigurationModel struct {
	AutoScaleEnabled                  types.Bool   `tfsdk:"auto_scale_enabled"`
	HostIPAddress                     types.String `tfsdk:"host_ip_address"`
	PrefixLength                      types.Int64  `tfsdk:"prefix_length"`
	AdditionalHostIPSubnets           types.List   `tfsdk:"additional_host_ip_subnets"`
	DHCPConfiguration                 types.Object `tfsdk:"dhcp_configuration"`
	NatOutboundIPAddressConfiguration types.List   `tfsdk:"nat_outbound_ip_address_configuration"`
}

type IPv6AddressSuffixRangeModel struct {
	Start types.String `tfsdk:"start"`
	Stop  types.String `tfsdk:"stop"`
}

type IPv6DHCPConfigurationModel struct {
	IPAddressSuffixRange types.Object `tfsdk:"ip_address_suffix_range"`
	LeaseTimeSeconds     types.Int64  `tfsdk:"lease_time_seconds"`
}

type IPv6ClientAddressAssignmentModel struct {
	DHCPConfiguration types.Object `tfsdk:"dhcp_configuration"`
	SlaacEnabled      types.Bool   `tfsdk:"slaac_enabled"`
}

type IPv6RouterAdvertisementModel struct {
	Priority types.String `tfsdk:"priority"`
}

type NetworkIPv6ConfigurationModel struct {
	InterfaceType                  types.String `tfsdk:"interface_type"`
	ClientAddressAssignment        types.Object `tfsdk:"client_address_assignment"`
	RouterAdvertisement            types.Object `tfsdk:"router_advertisement"`
	DNSServerIPAddressesOverride   types.List   `tfsdk:"dns_server_ip_addresses_override"`
	AdditionalHostIPSubnets        types.List   `tfsdk:"additional_host_ip_subnets"`
	PrefixDelegationWanInterfaceID types.String `tfsdk:"prefix_delegation_wan_interface_id"`
	HostIPAddress                  types.String `tfsdk:"host_ip_address"`
	PrefixLength                   types.String `tfsdk:"prefix_length"`
}

type DHCPGuardingModel struct {
	TrustedDHCPServerIPAddresses types.List `tfsdk:"trusted_dhcp_server_ip_addresses"`
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
	CellularBackupEnabled types.Bool   `tfsdk:"cellular_backup_enabled"`
	DeviceID              types.String `tfsdk:"device_id"`
	ZoneID                types.String `tfsdk:"zone_id"`
	DHCPGuarding          types.Object `tfsdk:"dhcp_guarding"`
	IPv4Configuration     types.Object `tfsdk:"ipv4_configuration"`
	IPv6Configuration     types.Object `tfsdk:"ipv6_configuration"`
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
			"cellular_backup_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether cellular backup is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"device_id": schema.StringAttribute{
				MarkdownDescription: "The device ID associated with this network.",
				Optional:            true,
			},
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "The firewall zone ID for this network.",
				Optional:            true,
			},
			"dhcp_guarding": schema.SingleNestedAttribute{
				MarkdownDescription: "DHCP guarding configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"trusted_dhcp_server_ip_addresses": schema.ListAttribute{
						MarkdownDescription: "List of trusted DHCP server IP addresses.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"ipv4_configuration": schema.SingleNestedAttribute{
				MarkdownDescription: "IPv4 configuration for the network.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"auto_scale_enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether auto-scaling is enabled.",
						Optional:            true,
					},
					"host_ip_address": schema.StringAttribute{
						MarkdownDescription: "The host IP address (gateway).",
						Optional:            true,
					},
					"prefix_length": schema.Int64Attribute{
						MarkdownDescription: "The prefix length (subnet mask).",
						Optional:            true,
					},
					"additional_host_ip_subnets": schema.ListAttribute{
						MarkdownDescription: "Additional host IP subnets.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"dhcp_configuration": schema.SingleNestedAttribute{
						MarkdownDescription: "DHCP configuration.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								MarkdownDescription: "DHCP mode (dhcp-server, dhcp-relay, none).",
								Required:            true,
							},
							"ip_address_range": schema.SingleNestedAttribute{
								MarkdownDescription: "DHCP IP address range.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"start": schema.StringAttribute{
										MarkdownDescription: "Start IP address.",
										Optional:            true,
									},
									"stop": schema.StringAttribute{
										MarkdownDescription: "Stop IP address.",
										Optional:            true,
									},
								},
							},
							"gateway_ip_address_override": schema.StringAttribute{
								MarkdownDescription: "Gateway IP address override.",
								Optional:            true,
							},
							"dns_server_ip_addresses_override": schema.ListAttribute{
								MarkdownDescription: "DNS server IP addresses override.",
								Optional:            true,
								ElementType:         types.StringType,
							},
							"lease_time_seconds": schema.Int64Attribute{
								MarkdownDescription: "DHCP lease time in seconds.",
								Optional:            true,
							},
							"domain_name": schema.StringAttribute{
								MarkdownDescription: "Domain name for DHCP clients.",
								Optional:            true,
							},
							"ping_conflict_detection_enabled": schema.BoolAttribute{
								MarkdownDescription: "Whether ping conflict detection is enabled.",
								Optional:            true,
							},
							"pxe_configuration": schema.SingleNestedAttribute{
								MarkdownDescription: "PXE boot configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"server_ip_address": schema.StringAttribute{
										MarkdownDescription: "PXE server IP address.",
										Required:            true,
									},
									"filename": schema.StringAttribute{
										MarkdownDescription: "PXE boot filename.",
										Required:            true,
									},
								},
							},
							"ntp_server_ip_addresses": schema.ListAttribute{
								MarkdownDescription: "NTP server IP addresses.",
								Optional:            true,
								ElementType:         types.StringType,
							},
							"option43_value": schema.StringAttribute{
								MarkdownDescription: "DHCP option 43 value.",
								Optional:            true,
							},
							"tftp_server_address": schema.StringAttribute{
								MarkdownDescription: "TFTP server address.",
								Optional:            true,
							},
							"time_offset_seconds": schema.Int64Attribute{
								MarkdownDescription: "Time offset in seconds.",
								Optional:            true,
							},
							"wpad_url": schema.StringAttribute{
								MarkdownDescription: "WPAD URL.",
								Optional:            true,
							},
							"wins_server_ip_addresses": schema.ListAttribute{
								MarkdownDescription: "WINS server IP addresses.",
								Optional:            true,
								ElementType:         types.StringType,
							},
							"dhcp_server_ip_addresses": schema.ListAttribute{
								MarkdownDescription: "DHCP server IP addresses (for relay mode).",
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"nat_outbound_ip_address_configuration": schema.ListNestedAttribute{
						MarkdownDescription: "NAT outbound IP address configuration.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "NAT type.",
									Required:            true,
								},
								"wan_interface_id": schema.StringAttribute{
									MarkdownDescription: "WAN interface ID.",
									Required:            true,
								},
								"ip_address_selectors": schema.ListNestedAttribute{
									MarkdownDescription: "IP address selectors.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												MarkdownDescription: "Selector type.",
												Required:            true,
											},
											"value": schema.StringAttribute{
												MarkdownDescription: "Selector value.",
												Optional:            true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"ipv6_configuration": schema.SingleNestedAttribute{
				MarkdownDescription: "IPv6 configuration for the network.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"interface_type": schema.StringAttribute{
						MarkdownDescription: "IPv6 interface type (static, prefix-delegation, none).",
						Required:            true,
					},
					"client_address_assignment": schema.SingleNestedAttribute{
						MarkdownDescription: "Client address assignment configuration.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"dhcp_configuration": schema.SingleNestedAttribute{
								MarkdownDescription: "DHCPv6 configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"ip_address_suffix_range": schema.SingleNestedAttribute{
										MarkdownDescription: "IPv6 address suffix range.",
										Optional:            true,
										Attributes: map[string]schema.Attribute{
											"start": schema.StringAttribute{
												MarkdownDescription: "Start suffix.",
												Optional:            true,
											},
											"stop": schema.StringAttribute{
												MarkdownDescription: "Stop suffix.",
												Optional:            true,
											},
										},
									},
									"lease_time_seconds": schema.Int64Attribute{
										MarkdownDescription: "DHCPv6 lease time in seconds.",
										Optional:            true,
									},
								},
							},
							"slaac_enabled": schema.BoolAttribute{
								MarkdownDescription: "Whether SLAAC is enabled.",
								Optional:            true,
							},
						},
					},
					"router_advertisement": schema.SingleNestedAttribute{
						MarkdownDescription: "Router advertisement configuration.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"priority": schema.StringAttribute{
								MarkdownDescription: "Router advertisement priority (high, medium, low).",
								Optional:            true,
							},
						},
					},
					"dns_server_ip_addresses_override": schema.ListAttribute{
						MarkdownDescription: "DNS server IPv6 addresses override.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"additional_host_ip_subnets": schema.ListAttribute{
						MarkdownDescription: "Additional host IPv6 subnets.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"prefix_delegation_wan_interface_id": schema.StringAttribute{
						MarkdownDescription: "WAN interface ID for prefix delegation.",
						Optional:            true,
					},
					"host_ip_address": schema.StringAttribute{
						MarkdownDescription: "Host IPv6 address.",
						Optional:            true,
					},
					"prefix_length": schema.StringAttribute{
						MarkdownDescription: "IPv6 prefix length.",
						Optional:            true,
					},
				},
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

	createReq := r.buildCreateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

	r.mapResponseToModel(ctx, networkResp, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

	updateReq := r.buildUpdateRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

func (r *NetworkResource) buildCreateRequest(ctx context.Context, data *NetworkResourceModel, diags *diag.Diagnostics) networktypes.CreateNetworkRequest {
	isolationEnabled := data.IsolationEnabled.ValueBool()
	internetAccessEnabled := data.InternetAccessEnabled.ValueBool()
	mdnsForwardingEnabled := data.MdnsForwardingEnabled.ValueBool()
	cellularBackupEnabled := data.CellularBackupEnabled.ValueBool()

	createReq := networktypes.CreateNetworkRequest{
		SiteID:                data.SiteID.ValueString(),
		Name:                  data.Name.ValueString(),
		Enabled:               data.Enabled.ValueBool(),
		VlanID:                int(data.VlanID.ValueInt64()),
		Management:            data.Management.ValueString(),
		IsolationEnabled:      &isolationEnabled,
		InternetAccessEnabled: &internetAccessEnabled,
		MdnsForwardingEnabled: &mdnsForwardingEnabled,
		CellularBackupEnabled: &cellularBackupEnabled,
		DeviceID:              data.DeviceID.ValueString(),
		ZoneID:                data.ZoneID.ValueString(),
	}

	if !data.DHCPGuarding.IsNull() && !data.DHCPGuarding.IsUnknown() {
		var dhcpGuarding DHCPGuardingModel
		diags.Append(data.DHCPGuarding.As(ctx, &dhcpGuarding, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			var trustedIPs []string
			diags.Append(dhcpGuarding.TrustedDHCPServerIPAddresses.ElementsAs(ctx, &trustedIPs, false)...)
			createReq.DHCPGuarding = &networktypes.DHCPGuarding{
				TrustedDHCPServerIPAddresses: trustedIPs,
			}
		}
	}

	if !data.IPv4Configuration.IsNull() && !data.IPv4Configuration.IsUnknown() {
		createReq.IPv4Configuration = r.buildIPv4Configuration(ctx, data.IPv4Configuration, diags)
	}

	if !data.IPv6Configuration.IsNull() && !data.IPv6Configuration.IsUnknown() {
		createReq.IPv6Configuration = r.buildIPv6Configuration(ctx, data.IPv6Configuration, diags)
	}

	return createReq
}

func (r *NetworkResource) buildUpdateRequest(ctx context.Context, data *NetworkResourceModel, diags *diag.Diagnostics) networktypes.UpdateNetworkRequest {
	isolationEnabled := data.IsolationEnabled.ValueBool()
	internetAccessEnabled := data.InternetAccessEnabled.ValueBool()
	mdnsForwardingEnabled := data.MdnsForwardingEnabled.ValueBool()
	cellularBackupEnabled := data.CellularBackupEnabled.ValueBool()

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
		CellularBackupEnabled: &cellularBackupEnabled,
		DeviceID:              data.DeviceID.ValueString(),
		ZoneID:                data.ZoneID.ValueString(),
	}

	if !data.DHCPGuarding.IsNull() && !data.DHCPGuarding.IsUnknown() {
		var dhcpGuarding DHCPGuardingModel
		diags.Append(data.DHCPGuarding.As(ctx, &dhcpGuarding, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			var trustedIPs []string
			diags.Append(dhcpGuarding.TrustedDHCPServerIPAddresses.ElementsAs(ctx, &trustedIPs, false)...)
			updateReq.DHCPGuarding = &networktypes.DHCPGuarding{
				TrustedDHCPServerIPAddresses: trustedIPs,
			}
		}
	}

	if !data.IPv4Configuration.IsNull() && !data.IPv4Configuration.IsUnknown() {
		updateReq.IPv4Configuration = r.buildIPv4Configuration(ctx, data.IPv4Configuration, diags)
	}

	if !data.IPv6Configuration.IsNull() && !data.IPv6Configuration.IsUnknown() {
		updateReq.IPv6Configuration = r.buildIPv6Configuration(ctx, data.IPv6Configuration, diags)
	}

	return updateReq
}

func (r *NetworkResource) buildIPv4Configuration(ctx context.Context, ipv4Obj types.Object, diags *diag.Diagnostics) *networktypes.NetworkIPv4Configuration {
	var ipv4Config NetworkIPv4ConfigurationModel
	diags.Append(ipv4Obj.As(ctx, &ipv4Config, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.NetworkIPv4Configuration{
		HostIPAddress: ipv4Config.HostIPAddress.ValueString(),
	}

	if !ipv4Config.AutoScaleEnabled.IsNull() {
		autoScale := ipv4Config.AutoScaleEnabled.ValueBool()
		result.AutoScaleEnabled = &autoScale
	}

	if !ipv4Config.PrefixLength.IsNull() {
		prefixLen := int(ipv4Config.PrefixLength.ValueInt64())
		result.PrefixLength = &prefixLen
	}

	if !ipv4Config.AdditionalHostIPSubnets.IsNull() {
		var subnets []string
		diags.Append(ipv4Config.AdditionalHostIPSubnets.ElementsAs(ctx, &subnets, false)...)
		result.AdditionalHostIPSubnets = subnets
	}

	if !ipv4Config.DHCPConfiguration.IsNull() && !ipv4Config.DHCPConfiguration.IsUnknown() {
		result.DHCPConfiguration = r.buildDHCPConfiguration(ctx, ipv4Config.DHCPConfiguration, diags)
	}

	if !ipv4Config.NatOutboundIPAddressConfiguration.IsNull() {
		result.NatOutboundIPAddressConfiguration = r.buildNATOutboundConfig(ctx, ipv4Config.NatOutboundIPAddressConfiguration, diags)
	}

	return result
}

func (r *NetworkResource) buildDHCPConfiguration(ctx context.Context, dhcpObj types.Object, diags *diag.Diagnostics) *networktypes.NetworkDHCPConfiguration {
	var dhcpConfig NetworkDHCPConfigurationModel
	diags.Append(dhcpObj.As(ctx, &dhcpConfig, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.NetworkDHCPConfiguration{
		Mode:                     dhcpConfig.Mode.ValueString(),
		GatewayIPAddressOverride: dhcpConfig.GatewayIPAddressOverride.ValueString(),
		DomainName:               dhcpConfig.DomainName.ValueString(),
		Option43Value:            dhcpConfig.Option43Value.ValueString(),
		TftpServerAddress:        dhcpConfig.TftpServerAddress.ValueString(),
		WpadURL:                  dhcpConfig.WpadURL.ValueString(),
	}

	if !dhcpConfig.IPAddressRange.IsNull() && !dhcpConfig.IPAddressRange.IsUnknown() {
		var ipRange NetworkDHCPIPAddressRangeModel
		diags.Append(dhcpConfig.IPAddressRange.As(ctx, &ipRange, basetypes.ObjectAsOptions{})...)
		result.IPAddressRange = &networktypes.NetworkDHCPIPAddressRange{
			Start: ipRange.Start.ValueString(),
			Stop:  ipRange.Stop.ValueString(),
		}
	}

	if !dhcpConfig.DNSServerIPAddressesOverride.IsNull() {
		var dnsServers []string
		diags.Append(dhcpConfig.DNSServerIPAddressesOverride.ElementsAs(ctx, &dnsServers, false)...)
		result.DNSServerIPAddressesOverride = dnsServers
	}

	if !dhcpConfig.LeaseTimeSeconds.IsNull() {
		leaseTime := int(dhcpConfig.LeaseTimeSeconds.ValueInt64())
		result.LeaseTimeSeconds = &leaseTime
	}

	if !dhcpConfig.PingConflictDetectionEnabled.IsNull() {
		pingDetect := dhcpConfig.PingConflictDetectionEnabled.ValueBool()
		result.PingConflictDetectionEnabled = &pingDetect
	}

	if !dhcpConfig.PxeConfiguration.IsNull() && !dhcpConfig.PxeConfiguration.IsUnknown() {
		var pxeConfig NetworkPXEConfigurationModel
		diags.Append(dhcpConfig.PxeConfiguration.As(ctx, &pxeConfig, basetypes.ObjectAsOptions{})...)
		result.PxeConfiguration = &networktypes.NetworkPXEConfiguration{
			ServerIPAddress: pxeConfig.ServerIPAddress.ValueString(),
			Filename:        pxeConfig.Filename.ValueString(),
		}
	}

	if !dhcpConfig.NtpServerIPAddresses.IsNull() {
		var ntpServers []string
		diags.Append(dhcpConfig.NtpServerIPAddresses.ElementsAs(ctx, &ntpServers, false)...)
		result.NtpServerIPAddresses = ntpServers
	}

	if !dhcpConfig.TimeOffsetSeconds.IsNull() {
		timeOffset := int(dhcpConfig.TimeOffsetSeconds.ValueInt64())
		result.TimeOffsetSeconds = &timeOffset
	}

	if !dhcpConfig.WinsServerIPAddresses.IsNull() {
		var winsServers []string
		diags.Append(dhcpConfig.WinsServerIPAddresses.ElementsAs(ctx, &winsServers, false)...)
		result.WinsServerIPAddresses = winsServers
	}

	if !dhcpConfig.DHCPServerIPAddresses.IsNull() {
		var dhcpServers []string
		diags.Append(dhcpConfig.DHCPServerIPAddresses.ElementsAs(ctx, &dhcpServers, false)...)
		result.DHCPServerIPAddresses = dhcpServers
	}

	return result
}

func (r *NetworkResource) buildNATOutboundConfig(ctx context.Context, natList types.List, diags *diag.Diagnostics) []networktypes.NetworkNATOutboundIPAddressConfig {
	var natConfigs []NetworkNATOutboundIPAddressConfigModel
	diags.Append(natList.ElementsAs(ctx, &natConfigs, false)...)
	if diags.HasError() {
		return nil
	}

	var result []networktypes.NetworkNATOutboundIPAddressConfig
	for _, natConfig := range natConfigs {
		config := networktypes.NetworkNATOutboundIPAddressConfig{
			Type:           natConfig.Type.ValueString(),
			WanInterfaceID: natConfig.WanInterfaceID.ValueString(),
		}

		if !natConfig.IpAddressSelectors.IsNull() {
			var selectors []IPAddressSelectorModel
			diags.Append(natConfig.IpAddressSelectors.ElementsAs(ctx, &selectors, false)...)
			for _, sel := range selectors {
				config.IpAddressSelectors = append(config.IpAddressSelectors, networktypes.IPAddressSelector{
					Type:  sel.Type.ValueString(),
					Value: sel.Value.ValueString(),
				})
			}
		}

		result = append(result, config)
	}

	return result
}

func (r *NetworkResource) buildIPv6Configuration(ctx context.Context, ipv6Obj types.Object, diags *diag.Diagnostics) *networktypes.NetworkIPv6Configuration {
	var ipv6Config NetworkIPv6ConfigurationModel
	diags.Append(ipv6Obj.As(ctx, &ipv6Config, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := &networktypes.NetworkIPv6Configuration{
		InterfaceType:                  ipv6Config.InterfaceType.ValueString(),
		PrefixDelegationWanInterfaceID: ipv6Config.PrefixDelegationWanInterfaceID.ValueString(),
		HostIPAddress:                  ipv6Config.HostIPAddress.ValueString(),
		PrefixLength:                   ipv6Config.PrefixLength.ValueString(),
	}

	if !ipv6Config.ClientAddressAssignment.IsNull() && !ipv6Config.ClientAddressAssignment.IsUnknown() {
		var clientAssignment IPv6ClientAddressAssignmentModel
		diags.Append(ipv6Config.ClientAddressAssignment.As(ctx, &clientAssignment, basetypes.ObjectAsOptions{})...)

		result.ClientAddressAssignment = &networktypes.IPv6ClientAddressAssignment{
			SlaacEnabled: clientAssignment.SlaacEnabled.ValueBool(),
		}

		if !clientAssignment.DHCPConfiguration.IsNull() && !clientAssignment.DHCPConfiguration.IsUnknown() {
			var dhcpv6Config IPv6DHCPConfigurationModel
			diags.Append(clientAssignment.DHCPConfiguration.As(ctx, &dhcpv6Config, basetypes.ObjectAsOptions{})...)

			result.ClientAddressAssignment.DHCPConfiguration = &networktypes.IPv6DHCPConfiguration{
				LeaseTimeSeconds: int(dhcpv6Config.LeaseTimeSeconds.ValueInt64()),
			}

			if !dhcpv6Config.IPAddressSuffixRange.IsNull() && !dhcpv6Config.IPAddressSuffixRange.IsUnknown() {
				var suffixRange IPv6AddressSuffixRangeModel
				diags.Append(dhcpv6Config.IPAddressSuffixRange.As(ctx, &suffixRange, basetypes.ObjectAsOptions{})...)
				result.ClientAddressAssignment.DHCPConfiguration.IPAddressSuffixRange = &networktypes.IPv6AddressSuffixRange{
					Start: suffixRange.Start.ValueString(),
					Stop:  suffixRange.Stop.ValueString(),
				}
			}
		}
	}

	if !ipv6Config.RouterAdvertisement.IsNull() && !ipv6Config.RouterAdvertisement.IsUnknown() {
		var ra IPv6RouterAdvertisementModel
		diags.Append(ipv6Config.RouterAdvertisement.As(ctx, &ra, basetypes.ObjectAsOptions{})...)
		result.RouterAdvertisement = &networktypes.IPv6RouterAdvertisement{
			Priority: ra.Priority.ValueString(),
		}
	}

	if !ipv6Config.DNSServerIPAddressesOverride.IsNull() {
		var dnsServers []string
		diags.Append(ipv6Config.DNSServerIPAddressesOverride.ElementsAs(ctx, &dnsServers, false)...)
		result.DNSServerIPAddressesOverride = dnsServers
	}

	if !ipv6Config.AdditionalHostIPSubnets.IsNull() {
		var subnets []string
		diags.Append(ipv6Config.AdditionalHostIPSubnets.ElementsAs(ctx, &subnets, false)...)
		result.AdditionalHostIPSubnets = subnets
	}

	return result
}

func (r *NetworkResource) mapResponseToModel(ctx context.Context, resp *networktypes.Network, data *NetworkResourceModel, diags *diag.Diagnostics) {
	data.Name = types.StringValue(resp.Name)
	data.Enabled = types.BoolValue(resp.Enabled)
	data.VlanID = types.Int64Value(int64(resp.VlanID))
	data.Management = types.StringValue(resp.Management)
	data.DeviceID = types.StringValue(resp.DeviceID)
	data.ZoneID = types.StringValue(resp.ZoneID)

	if resp.IsolationEnabled != nil {
		data.IsolationEnabled = types.BoolValue(*resp.IsolationEnabled)
	}
	if resp.InternetAccessEnabled != nil {
		data.InternetAccessEnabled = types.BoolValue(*resp.InternetAccessEnabled)
	}
	if resp.MdnsForwardingEnabled != nil {
		data.MdnsForwardingEnabled = types.BoolValue(*resp.MdnsForwardingEnabled)
	}
	if resp.CellularBackupEnabled != nil {
		data.CellularBackupEnabled = types.BoolValue(*resp.CellularBackupEnabled)
	}

	if resp.DHCPGuarding != nil {
		trustedIPs, d := types.ListValueFrom(ctx, types.StringType, resp.DHCPGuarding.TrustedDHCPServerIPAddresses)
		diags.Append(d...)
		dhcpGuardingObj, d := types.ObjectValue(
			map[string]attr.Type{"trusted_dhcp_server_ip_addresses": types.ListType{ElemType: types.StringType}},
			map[string]attr.Value{"trusted_dhcp_server_ip_addresses": trustedIPs},
		)
		diags.Append(d...)
		data.DHCPGuarding = dhcpGuardingObj
	}

	if resp.IPv4Configuration != nil {
		data.IPv4Configuration = r.mapIPv4ConfigurationToObject(ctx, resp.IPv4Configuration, diags)
	}

	if resp.IPv6Configuration != nil {
		data.IPv6Configuration = r.mapIPv6ConfigurationToObject(ctx, resp.IPv6Configuration, diags)
	}
}

func (r *NetworkResource) mapIPv4ConfigurationToObject(ctx context.Context, ipv4 *networktypes.NetworkIPv4Configuration, diags *diag.Diagnostics) types.Object {
	attrTypes := map[string]attr.Type{
		"auto_scale_enabled":                    types.BoolType,
		"host_ip_address":                       types.StringType,
		"prefix_length":                         types.Int64Type,
		"additional_host_ip_subnets":            types.ListType{ElemType: types.StringType},
		"dhcp_configuration":                    types.ObjectType{AttrTypes: getDHCPConfigAttrTypes()},
		"nat_outbound_ip_address_configuration": types.ListType{ElemType: types.ObjectType{AttrTypes: getNATOutboundAttrTypes()}},
	}

	attrValues := map[string]attr.Value{
		"host_ip_address": types.StringValue(ipv4.HostIPAddress),
	}

	if ipv4.AutoScaleEnabled != nil {
		attrValues["auto_scale_enabled"] = types.BoolValue(*ipv4.AutoScaleEnabled)
	} else {
		attrValues["auto_scale_enabled"] = types.BoolNull()
	}

	if ipv4.PrefixLength != nil {
		attrValues["prefix_length"] = types.Int64Value(int64(*ipv4.PrefixLength))
	} else {
		attrValues["prefix_length"] = types.Int64Null()
	}

	if len(ipv4.AdditionalHostIPSubnets) > 0 {
		subnets, d := types.ListValueFrom(ctx, types.StringType, ipv4.AdditionalHostIPSubnets)
		diags.Append(d...)
		attrValues["additional_host_ip_subnets"] = subnets
	} else {
		attrValues["additional_host_ip_subnets"] = types.ListNull(types.StringType)
	}

	if ipv4.DHCPConfiguration != nil {
		attrValues["dhcp_configuration"] = r.mapDHCPConfigToObject(ctx, ipv4.DHCPConfiguration, diags)
	} else {
		attrValues["dhcp_configuration"] = types.ObjectNull(getDHCPConfigAttrTypes())
	}

	if len(ipv4.NatOutboundIPAddressConfiguration) > 0 {
		attrValues["nat_outbound_ip_address_configuration"] = r.mapNATOutboundToList(ctx, ipv4.NatOutboundIPAddressConfiguration, diags)
	} else {
		attrValues["nat_outbound_ip_address_configuration"] = types.ListNull(types.ObjectType{AttrTypes: getNATOutboundAttrTypes()})
	}

	obj, d := types.ObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	return obj
}

func getDHCPConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode": types.StringType,
		"ip_address_range": types.ObjectType{AttrTypes: map[string]attr.Type{
			"start": types.StringType,
			"stop":  types.StringType,
		}},
		"gateway_ip_address_override":      types.StringType,
		"dns_server_ip_addresses_override": types.ListType{ElemType: types.StringType},
		"lease_time_seconds":               types.Int64Type,
		"domain_name":                      types.StringType,
		"ping_conflict_detection_enabled":  types.BoolType,
		"pxe_configuration": types.ObjectType{AttrTypes: map[string]attr.Type{
			"server_ip_address": types.StringType,
			"filename":          types.StringType,
		}},
		"ntp_server_ip_addresses":  types.ListType{ElemType: types.StringType},
		"option43_value":           types.StringType,
		"tftp_server_address":      types.StringType,
		"time_offset_seconds":      types.Int64Type,
		"wpad_url":                 types.StringType,
		"wins_server_ip_addresses": types.ListType{ElemType: types.StringType},
		"dhcp_server_ip_addresses": types.ListType{ElemType: types.StringType},
	}
}

func getNATOutboundAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":             types.StringType,
		"wan_interface_id": types.StringType,
		"ip_address_selectors": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":  types.StringType,
			"value": types.StringType,
		}}},
	}
}

func (r *NetworkResource) mapDHCPConfigToObject(ctx context.Context, dhcp *networktypes.NetworkDHCPConfiguration, diags *diag.Diagnostics) types.Object {
	attrValues := map[string]attr.Value{
		"mode":                        types.StringValue(dhcp.Mode),
		"gateway_ip_address_override": types.StringValue(dhcp.GatewayIPAddressOverride),
		"domain_name":                 types.StringValue(dhcp.DomainName),
		"option43_value":              types.StringValue(dhcp.Option43Value),
		"tftp_server_address":         types.StringValue(dhcp.TftpServerAddress),
		"wpad_url":                    types.StringValue(dhcp.WpadURL),
	}

	if dhcp.IPAddressRange != nil {
		ipRangeObj, d := types.ObjectValue(
			map[string]attr.Type{"start": types.StringType, "stop": types.StringType},
			map[string]attr.Value{
				"start": types.StringValue(dhcp.IPAddressRange.Start),
				"stop":  types.StringValue(dhcp.IPAddressRange.Stop),
			},
		)
		diags.Append(d...)
		attrValues["ip_address_range"] = ipRangeObj
	} else {
		attrValues["ip_address_range"] = types.ObjectNull(map[string]attr.Type{"start": types.StringType, "stop": types.StringType})
	}

	if len(dhcp.DNSServerIPAddressesOverride) > 0 {
		dnsServers, d := types.ListValueFrom(ctx, types.StringType, dhcp.DNSServerIPAddressesOverride)
		diags.Append(d...)
		attrValues["dns_server_ip_addresses_override"] = dnsServers
	} else {
		attrValues["dns_server_ip_addresses_override"] = types.ListNull(types.StringType)
	}

	if dhcp.LeaseTimeSeconds != nil {
		attrValues["lease_time_seconds"] = types.Int64Value(int64(*dhcp.LeaseTimeSeconds))
	} else {
		attrValues["lease_time_seconds"] = types.Int64Null()
	}

	if dhcp.PingConflictDetectionEnabled != nil {
		attrValues["ping_conflict_detection_enabled"] = types.BoolValue(*dhcp.PingConflictDetectionEnabled)
	} else {
		attrValues["ping_conflict_detection_enabled"] = types.BoolNull()
	}

	if dhcp.PxeConfiguration != nil {
		pxeObj, d := types.ObjectValue(
			map[string]attr.Type{"server_ip_address": types.StringType, "filename": types.StringType},
			map[string]attr.Value{
				"server_ip_address": types.StringValue(dhcp.PxeConfiguration.ServerIPAddress),
				"filename":          types.StringValue(dhcp.PxeConfiguration.Filename),
			},
		)
		diags.Append(d...)
		attrValues["pxe_configuration"] = pxeObj
	} else {
		attrValues["pxe_configuration"] = types.ObjectNull(map[string]attr.Type{"server_ip_address": types.StringType, "filename": types.StringType})
	}

	if len(dhcp.NtpServerIPAddresses) > 0 {
		ntpServers, d := types.ListValueFrom(ctx, types.StringType, dhcp.NtpServerIPAddresses)
		diags.Append(d...)
		attrValues["ntp_server_ip_addresses"] = ntpServers
	} else {
		attrValues["ntp_server_ip_addresses"] = types.ListNull(types.StringType)
	}

	if dhcp.TimeOffsetSeconds != nil {
		attrValues["time_offset_seconds"] = types.Int64Value(int64(*dhcp.TimeOffsetSeconds))
	} else {
		attrValues["time_offset_seconds"] = types.Int64Null()
	}

	if len(dhcp.WinsServerIPAddresses) > 0 {
		winsServers, d := types.ListValueFrom(ctx, types.StringType, dhcp.WinsServerIPAddresses)
		diags.Append(d...)
		attrValues["wins_server_ip_addresses"] = winsServers
	} else {
		attrValues["wins_server_ip_addresses"] = types.ListNull(types.StringType)
	}

	if len(dhcp.DHCPServerIPAddresses) > 0 {
		dhcpServers, d := types.ListValueFrom(ctx, types.StringType, dhcp.DHCPServerIPAddresses)
		diags.Append(d...)
		attrValues["dhcp_server_ip_addresses"] = dhcpServers
	} else {
		attrValues["dhcp_server_ip_addresses"] = types.ListNull(types.StringType)
	}

	obj, d := types.ObjectValue(getDHCPConfigAttrTypes(), attrValues)
	diags.Append(d...)
	return obj
}

func (r *NetworkResource) mapNATOutboundToList(ctx context.Context, natConfigs []networktypes.NetworkNATOutboundIPAddressConfig, diags *diag.Diagnostics) types.List {
	var elements []attr.Value
	for _, nat := range natConfigs {
		var selectors []attr.Value
		for _, sel := range nat.IpAddressSelectors {
			selObj, d := types.ObjectValue(
				map[string]attr.Type{"type": types.StringType, "value": types.StringType},
				map[string]attr.Value{"type": types.StringValue(sel.Type), "value": types.StringValue(sel.Value)},
			)
			diags.Append(d...)
			selectors = append(selectors, selObj)
		}

		selectorsList, d := types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{"type": types.StringType, "value": types.StringType}}, selectors)
		diags.Append(d...)

		natObj, d := types.ObjectValue(getNATOutboundAttrTypes(), map[string]attr.Value{
			"type":                 types.StringValue(nat.Type),
			"wan_interface_id":     types.StringValue(nat.WanInterfaceID),
			"ip_address_selectors": selectorsList,
		})
		diags.Append(d...)
		elements = append(elements, natObj)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: getNATOutboundAttrTypes()}, elements)
	diags.Append(d...)
	return list
}

func (r *NetworkResource) mapIPv6ConfigurationToObject(ctx context.Context, ipv6 *networktypes.NetworkIPv6Configuration, diags *diag.Diagnostics) types.Object {
	attrTypes := map[string]attr.Type{
		"interface_type": types.StringType,
		"client_address_assignment": types.ObjectType{AttrTypes: map[string]attr.Type{
			"dhcp_configuration": types.ObjectType{AttrTypes: map[string]attr.Type{
				"ip_address_suffix_range": types.ObjectType{AttrTypes: map[string]attr.Type{
					"start": types.StringType,
					"stop":  types.StringType,
				}},
				"lease_time_seconds": types.Int64Type,
			}},
			"slaac_enabled": types.BoolType,
		}},
		"router_advertisement": types.ObjectType{AttrTypes: map[string]attr.Type{
			"priority": types.StringType,
		}},
		"dns_server_ip_addresses_override":   types.ListType{ElemType: types.StringType},
		"additional_host_ip_subnets":         types.ListType{ElemType: types.StringType},
		"prefix_delegation_wan_interface_id": types.StringType,
		"host_ip_address":                    types.StringType,
		"prefix_length":                      types.StringType,
	}

	attrValues := map[string]attr.Value{
		"interface_type":                     types.StringValue(ipv6.InterfaceType),
		"prefix_delegation_wan_interface_id": types.StringValue(ipv6.PrefixDelegationWanInterfaceID),
		"host_ip_address":                    types.StringValue(ipv6.HostIPAddress),
		"prefix_length":                      types.StringValue(ipv6.PrefixLength),
	}

	if ipv6.ClientAddressAssignment != nil {
		clientAttrValues := map[string]attr.Value{
			"slaac_enabled": types.BoolValue(ipv6.ClientAddressAssignment.SlaacEnabled),
		}

		if ipv6.ClientAddressAssignment.DHCPConfiguration != nil {
			dhcpv6AttrValues := map[string]attr.Value{}
			if ipv6.ClientAddressAssignment.DHCPConfiguration.IPAddressSuffixRange != nil {
				suffixRangeObj, d := types.ObjectValue(
					map[string]attr.Type{"start": types.StringType, "stop": types.StringType},
					map[string]attr.Value{
						"start": types.StringValue(ipv6.ClientAddressAssignment.DHCPConfiguration.IPAddressSuffixRange.Start),
						"stop":  types.StringValue(ipv6.ClientAddressAssignment.DHCPConfiguration.IPAddressSuffixRange.Stop),
					},
				)
				diags.Append(d...)
				dhcpv6AttrValues["ip_address_suffix_range"] = suffixRangeObj
			} else {
				dhcpv6AttrValues["ip_address_suffix_range"] = types.ObjectNull(map[string]attr.Type{"start": types.StringType, "stop": types.StringType})
			}
			dhcpv6AttrValues["lease_time_seconds"] = types.Int64Value(int64(ipv6.ClientAddressAssignment.DHCPConfiguration.LeaseTimeSeconds))

			dhcpv6Obj, d := types.ObjectValue(
				map[string]attr.Type{
					"ip_address_suffix_range": types.ObjectType{AttrTypes: map[string]attr.Type{"start": types.StringType, "stop": types.StringType}},
					"lease_time_seconds":      types.Int64Type,
				},
				dhcpv6AttrValues,
			)
			diags.Append(d...)
			clientAttrValues["dhcp_configuration"] = dhcpv6Obj
		} else {
			clientAttrValues["dhcp_configuration"] = types.ObjectNull(map[string]attr.Type{
				"ip_address_suffix_range": types.ObjectType{AttrTypes: map[string]attr.Type{"start": types.StringType, "stop": types.StringType}},
				"lease_time_seconds":      types.Int64Type,
			})
		}

		clientObj, d := types.ObjectValue(
			map[string]attr.Type{
				"dhcp_configuration": types.ObjectType{AttrTypes: map[string]attr.Type{
					"ip_address_suffix_range": types.ObjectType{AttrTypes: map[string]attr.Type{"start": types.StringType, "stop": types.StringType}},
					"lease_time_seconds":      types.Int64Type,
				}},
				"slaac_enabled": types.BoolType,
			},
			clientAttrValues,
		)
		diags.Append(d...)
		attrValues["client_address_assignment"] = clientObj
	} else {
		attrValues["client_address_assignment"] = types.ObjectNull(map[string]attr.Type{
			"dhcp_configuration": types.ObjectType{AttrTypes: map[string]attr.Type{
				"ip_address_suffix_range": types.ObjectType{AttrTypes: map[string]attr.Type{"start": types.StringType, "stop": types.StringType}},
				"lease_time_seconds":      types.Int64Type,
			}},
			"slaac_enabled": types.BoolType,
		})
	}

	if ipv6.RouterAdvertisement != nil {
		raObj, d := types.ObjectValue(
			map[string]attr.Type{"priority": types.StringType},
			map[string]attr.Value{"priority": types.StringValue(ipv6.RouterAdvertisement.Priority)},
		)
		diags.Append(d...)
		attrValues["router_advertisement"] = raObj
	} else {
		attrValues["router_advertisement"] = types.ObjectNull(map[string]attr.Type{"priority": types.StringType})
	}

	if len(ipv6.DNSServerIPAddressesOverride) > 0 {
		dnsServers, d := types.ListValueFrom(ctx, types.StringType, ipv6.DNSServerIPAddressesOverride)
		diags.Append(d...)
		attrValues["dns_server_ip_addresses_override"] = dnsServers
	} else {
		attrValues["dns_server_ip_addresses_override"] = types.ListNull(types.StringType)
	}

	if len(ipv6.AdditionalHostIPSubnets) > 0 {
		subnets, d := types.ListValueFrom(ctx, types.StringType, ipv6.AdditionalHostIPSubnets)
		diags.Append(d...)
		attrValues["additional_host_ip_subnets"] = subnets
	} else {
		attrValues["additional_host_ip_subnets"] = types.ListNull(types.StringType)
	}

	obj, d := types.ObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	return obj
}
