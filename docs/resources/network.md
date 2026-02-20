---
page_title: "unifi_network Resource - UniFi Network"
subcategory: ""
description: |-
  Manages a UniFi network.
---

# unifi_network (Resource)

Manages a UniFi network.

## Example Usage

```terraform
resource "unifi_network" "guest" {
  site_id    = "your-site-id"
  name       = "Guest Network"
  enabled    = true
  vlan_id    = 100
  management = "third-party"

  isolation_enabled       = true
  internet_access_enabled = true
  mdns_forwarding_enabled = false
}
```

## Schema

### Required

- `site_id` (String) - The site ID where the network will be created.
- `name` (String) - The name of the network.

### Optional

- `enabled` (Boolean) - Whether the network is enabled. Defaults to `true`.
- `vlan_id` (Number) - The VLAN ID of the network. Defaults to `1`.
- `management` (String) - The management type of the network. Defaults to `third-party`.
- `isolation_enabled` (Boolean) - Whether network isolation is enabled. Defaults to `false`.
- `internet_access_enabled` (Boolean) - Whether internet access is enabled. Defaults to `true`.
- `mdns_forwarding_enabled` (Boolean) - Whether mDNS forwarding is enabled. Defaults to `false`.

### Read-Only

- `id` (String) - The unique identifier of the network.

## Import

Networks can be imported using the format `site_id:network_id`:

```shell
terraform import unifi_network.example site-id:network-id
```
