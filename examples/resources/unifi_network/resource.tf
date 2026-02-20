# Create a new network
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
