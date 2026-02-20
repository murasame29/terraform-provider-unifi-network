# Terraform Provider for UniFi Network

[![Go Report Card](https://goreportcard.com/badge/github.com/murasame29/terraform-provider-unifi-network)](https://goreportcard.com/report/github.com/murasame29/terraform-provider-unifi-network)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A Terraform provider for managing UniFi Network infrastructure using the [UniFi Cloud API](https://developer.ui.com/). This provider enables Infrastructure as Code (IaC) for UniFi networks, WiFi, firewall, DNS, ACL rules, and more.

## Features

- Full support for UniFi Cloud API
- Manage networks with IPv4/IPv6, DHCP, NAT configuration
- Configure WiFi broadcasts with WPA2/WPA3, band steering, MLO
- Create and manage firewall zones and policies
- Set up ACL rules for traffic control
- Configure DNS policies (A, AAAA, CNAME, MX, TXT, SRV, PTR records)
- Manage traffic matching lists
- Generate and manage hotspot vouchers

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- UniFi Cloud API key

## Installation

### From Terraform Registry (Coming Soon)

```terraform
terraform {
  required_providers {
    unifi = {
      source  = "murasame29/unifi-network"
      version = "~> 0.1"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/murasame29/terraform-provider-unifi-network.git
cd terraform-provider-unifi-network
go build -o terraform-provider-unifi-network
```

## Configuration

```terraform
provider "unifi" {
  api_key = var.unifi_api_key
  # base_url = "https://api.ui.com"  # Optional, defaults to UniFi Cloud API
}
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `UNIFI_API_KEY` | UniFi Cloud API key | Yes |
| `UNIFI_BASE_URL` | Base URL for the API | No |

## Resources

| Resource | Description |
|----------|-------------|
| `unifi_network` | Manage networks with VLAN, DHCP, IPv4/IPv6 configuration |
| `unifi_wifi_broadcast` | Manage WiFi broadcasts (SSIDs) with security settings |
| `unifi_firewall_zone` | Create firewall zones for network segmentation |
| `unifi_firewall_policy` | Define firewall policies between zones |
| `unifi_acl_rule` | Configure ACL rules for traffic control |
| `unifi_dns_policy` | Manage DNS records and policies |
| `unifi_traffic_matching_list` | Create traffic matching lists for firewall rules |
| `unifi_voucher` | Generate hotspot vouchers for guest access |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `unifi_sites` | List all sites |
| `unifi_network` | Get network details |
| `unifi_networks` | List networks for a site |
| `unifi_device` | Get device details |
| `unifi_devices` | List devices for a site |
| `unifi_clients` | List connected clients |
| `unifi_wifi_broadcasts` | List WiFi broadcasts |
| `unifi_firewall_zones` | List firewall zones |
| `unifi_firewall_policies` | List firewall policies |
| `unifi_acl_rules` | List ACL rules |
| `unifi_dns_policies` | List DNS policies |
| `unifi_traffic_matching_lists` | List traffic matching lists |
| `unifi_vouchers` | List hotspot vouchers |
| `unifi_wan_interfaces` | List WAN interfaces |
| `unifi_vpn_tunnels` | List VPN tunnels |
| `unifi_vpn_servers` | List VPN servers |
| `unifi_radius_profiles` | List RADIUS profiles |

## Example Usage

### Basic Network Setup

```terraform
provider "unifi" {
  api_key = var.unifi_api_key
}

# Get all sites
data "unifi_sites" "all" {}

# Create a corporate network with DHCP
resource "unifi_network" "corporate" {
  site_id    = data.unifi_sites.all.sites[0].id
  name       = "Corporate Network"
  enabled    = true
  vlan_id    = 10
  management = "third-party"

  ipv4_configuration = {
    static_subnet = "10.0.10.0/24"
    gateway       = "10.0.10.1"
    dns_servers   = ["8.8.8.8", "8.8.4.4"]
    dhcp = {
      mode               = "server"
      start              = "10.0.10.100"
      end                = "10.0.10.200"
      lease_time_seconds = 86400
    }
  }
}

# Create a guest network with isolation
resource "unifi_network" "guest" {
  site_id                 = data.unifi_sites.all.sites[0].id
  name                    = "Guest Network"
  enabled                 = true
  vlan_id                 = 100
  management              = "third-party"
  isolation_enabled       = true
  internet_access_enabled = true
}
```

### WiFi Configuration

```terraform
# Corporate WiFi with WPA3
resource "unifi_wifi_broadcast" "corporate" {
  site_id       = data.unifi_sites.all.sites[0].id
  name          = "Corporate WiFi"
  enabled       = true
  network_id    = unifi_network.corporate.id
  security_type = "wpa3"
  passphrase    = var.corporate_wifi_password

  security_configuration = {
    pmf_mode = "required"
  }

  band_steering = {
    mode = "prefer_5ghz"
  }
}

# Guest WiFi with client isolation
resource "unifi_wifi_broadcast" "guest" {
  site_id                  = data.unifi_sites.all.sites[0].id
  name                     = "Guest WiFi"
  enabled                  = true
  network_id               = unifi_network.guest.id
  security_type            = "wpa2"
  passphrase               = var.guest_wifi_password
  client_isolation_enabled = true
}
```

### Firewall Configuration

```terraform
# Create firewall zones
resource "unifi_firewall_zone" "internal" {
  site_id = data.unifi_sites.all.sites[0].id
  name    = "Internal"
}

resource "unifi_firewall_zone" "guest" {
  site_id = data.unifi_sites.all.sites[0].id
  name    = "Guest"
}

# Block guest from internal network
resource "unifi_firewall_policy" "block_guest_to_internal" {
  site_id = data.unifi_sites.all.sites[0].id
  name    = "Block Guest to Internal"
  enabled = true
  action  = "block"

  source_endpoint = {
    zone_id = unifi_firewall_zone.guest.id
  }

  destination_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }
}
```

### DNS Configuration

```terraform
# Internal DNS record
resource "unifi_dns_policy" "internal_server" {
  site_id = data.unifi_sites.all.sites[0].id
  name    = "Internal Server"
  enabled = true

  record = {
    type = "A"
    a = {
      hostname   = "server.local"
      ip_address = "10.0.10.50"
      ttl        = 3600
    }
  }
}
```

### Hotspot Vouchers

```terraform
# Generate guest vouchers
resource "unifi_voucher" "guest_day_pass" {
  site_id            = data.unifi_sites.all.sites[0].id
  name               = "Guest Day Pass"
  time_limit_minutes = 1440  # 24 hours
  voucher_count      = 10
  rx_rate_limit_kbps = 10000  # 10 Mbps download
  tx_rate_limit_kbps = 5000   # 5 Mbps upload
}
```

## Development

### Building

```bash
go build -o terraform-provider-unifi-network
```

### Testing

```bash
go test ./...
```

### Generating Documentation

```bash
go generate ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MPL-2.0 License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [unifi-client-go](https://github.com/murasame29/unifi-client-go) - Go client library for UniFi Cloud API
