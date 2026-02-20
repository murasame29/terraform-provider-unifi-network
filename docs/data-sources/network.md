---
page_title: "unifi_network Data Source - UniFi Network"
subcategory: ""
description: |-
  Fetches details of a specific UniFi network.
---

# unifi_network (Data Source)

Fetches details of a specific UniFi network.

## Example Usage

```terraform
data "unifi_network" "default" {
  site_id = "your-site-id"
  id      = "your-network-id"
}

output "network_vlan" {
  value = data.unifi_network.default.vlan_id
}
```

## Schema

### Required

- `site_id` (String) - The site ID where the network is located.
- `id` (String) - The unique identifier of the network.

### Read-Only

- `name` (String) - The name of the network.
- `enabled` (Boolean) - Whether the network is enabled.
- `vlan_id` (Number) - The VLAN ID of the network.
- `management` (String) - The management type of the network.
- `default` (Boolean) - Whether this is the default network.
- `isolation_enabled` (Boolean) - Whether network isolation is enabled.
- `internet_access_enabled` (Boolean) - Whether internet access is enabled.
