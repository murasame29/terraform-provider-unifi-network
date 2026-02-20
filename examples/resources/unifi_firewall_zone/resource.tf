# Basic firewall zone
resource "unifi_firewall_zone" "internal" {
  site_id     = "default"
  name        = "Internal Zone"
  description = "Internal trusted network zone"
}

# DMZ zone
resource "unifi_firewall_zone" "dmz" {
  site_id     = "default"
  name        = "DMZ"
  description = "Demilitarized zone for public-facing servers"
}

# Guest zone
resource "unifi_firewall_zone" "guest" {
  site_id     = "default"
  name        = "Guest Zone"
  description = "Isolated zone for guest networks"
}

# IoT zone
resource "unifi_firewall_zone" "iot" {
  site_id     = "default"
  name        = "IoT Zone"
  description = "Zone for IoT devices with restricted access"
}

# Management zone
resource "unifi_firewall_zone" "management" {
  site_id     = "default"
  name        = "Management"
  description = "Zone for network management devices"
}
