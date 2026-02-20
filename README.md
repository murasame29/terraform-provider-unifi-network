# Terraform Provider for UniFi Network

This Terraform provider allows you to manage UniFi Network resources using the [UniFi Cloud API](https://developer.ui.com/).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)

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

- `UNIFI_API_KEY` - UniFi Cloud API key
- `UNIFI_BASE_URL` - Base URL for the API (optional)

## Resources

- `unifi_network` - Manage networks
- `unifi_wifi_broadcast` - Manage WiFi broadcasts (SSIDs)

## Data Sources

- `unifi_sites` - List all sites
- `unifi_network` - Get network details
- `unifi_networks` - List networks for a site

## Example Usage

```terraform
provider "unifi" {
  api_key = var.unifi_api_key
}

# Get all sites
data "unifi_sites" "all" {}

# Create a guest network
resource "unifi_network" "guest" {
  site_id                 = data.unifi_sites.all.sites[0].id
  name                    = "Guest Network"
  enabled                 = true
  vlan_id                 = 100
  isolation_enabled       = true
  internet_access_enabled = true
}

# Create a WiFi broadcast for the guest network
resource "unifi_wifi_broadcast" "guest_wifi" {
  site_id       = data.unifi_sites.all.sites[0].id
  name          = "Guest WiFi"
  enabled       = true
  network_id    = unifi_network.guest.id
  security_type = "wpa2"
  passphrase    = var.wifi_passphrase
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

## License

MPL-2.0
