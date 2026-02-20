---
page_title: "UniFi Network Provider"
subcategory: ""
description: |-
  The UniFi Network provider allows you to manage UniFi Network resources using the UniFi Cloud API.
---

# UniFi Network Provider

The UniFi Network provider allows you to manage UniFi Network resources using the [UniFi Cloud API](https://developer.ui.com/).

## Example Usage

```terraform
provider "unifi" {
  api_key = var.unifi_api_key
}

# List all sites
data "unifi_sites" "all" {}

# Create a network
resource "unifi_network" "guest" {
  site_id    = data.unifi_sites.all.sites[0].id
  name       = "Guest Network"
  enabled    = true
  vlan_id    = 100
  management = "third-party"
}

# Create a WiFi broadcast
resource "unifi_wifi_broadcast" "guest_wifi" {
  site_id       = data.unifi_sites.all.sites[0].id
  name          = "Guest WiFi"
  enabled       = true
  network_id    = unifi_network.guest.id
  security_type = "wpa2"
  passphrase    = var.wifi_passphrase
}
```

## Authentication

The provider requires an API key from the UniFi Cloud. You can obtain an API key from the [UniFi Site Manager](https://unifi.ui.com/).

The API key can be provided in two ways:

1. Via the `api_key` attribute in the provider configuration
2. Via the `UNIFI_API_KEY` environment variable

## Schema

### Optional

- `api_key` (String, Sensitive) - The API key for authenticating with the UniFi Cloud API. Can also be set via the `UNIFI_API_KEY` environment variable.
- `base_url` (String) - The base URL for the UniFi Cloud API. Defaults to `https://api.ui.com`. Can also be set via the `UNIFI_BASE_URL` environment variable.
