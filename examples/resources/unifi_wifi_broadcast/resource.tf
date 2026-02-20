# Create a WiFi broadcast (SSID)
resource "unifi_wifi_broadcast" "guest_wifi" {
  site_id       = "your-site-id"
  name          = "Guest WiFi"
  enabled       = true
  network_id    = unifi_network.guest.id
  security_type = "wpa2"
  passphrase    = "secure-password-here"

  hide_name                = false
  client_isolation_enabled = true
}
