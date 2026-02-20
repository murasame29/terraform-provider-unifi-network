---
page_title: "unifi_wifi_broadcast Resource - UniFi Network"
subcategory: ""
description: |-
  Manages a UniFi WiFi broadcast (SSID).
---

# unifi_wifi_broadcast (Resource)

Manages a UniFi WiFi broadcast (SSID).

## Example Usage

```terraform
resource "unifi_wifi_broadcast" "guest_wifi" {
  site_id       = "your-site-id"
  name          = "Guest WiFi"
  enabled       = true
  network_id    = unifi_network.guest.id
  security_type = "wpa2"
  passphrase    = var.wifi_passphrase

  hide_name                = false
  client_isolation_enabled = true
}
```

## Schema

### Required

- `site_id` (String) - The site ID where the WiFi broadcast will be created.
- `name` (String) - The name (SSID) of the WiFi broadcast.

### Optional

- `type` (String) - The type of WiFi broadcast. Defaults to `standard`.
- `enabled` (Boolean) - Whether the WiFi broadcast is enabled. Defaults to `true`.
- `network_id` (String) - The network ID to associate with this WiFi broadcast.
- `security_type` (String) - The security type. Valid values: `open`, `wpa2`, `wpa3`, `wpa2wpa3`. Defaults to `wpa2`.
- `passphrase` (String, Sensitive) - The WiFi passphrase. Required when security_type is not `open`.
- `hide_name` (Boolean) - Whether to hide the SSID. Defaults to `false`.
- `client_isolation_enabled` (Boolean) - Whether client isolation is enabled. Defaults to `false`.

### Read-Only

- `id` (String) - The unique identifier of the WiFi broadcast.

## Import

WiFi broadcasts can be imported using the format `site_id:wifi_broadcast_id`:

```shell
terraform import unifi_wifi_broadcast.example site-id:wifi-broadcast-id
```
