---
page_title: "unifi_networks Data Source - UniFi Network"
subcategory: ""
description: |-
  Fetches the list of networks for a UniFi site.
---

# unifi_networks (Data Source)

Fetches the list of networks for a UniFi site.

## Example Usage

```terraform
data "unifi_networks" "all" {
  site_id = "your-site-id"
}

output "network_count" {
  value = length(data.unifi_networks.all.networks)
}
```

## Schema

### Required

- `site_id` (String) - The site ID to list networks for.

### Read-Only

- `networks` (List of Object) - List of networks.
  - `id` (String) - The unique identifier of the network.
  - `name` (String) - The name of the network.
  - `enabled` (Boolean) - Whether the network is enabled.
  - `vlan_id` (Number) - The VLAN ID of the network.
  - `management` (String) - The management type of the network.
  - `default` (Boolean) - Whether this is the default network.
