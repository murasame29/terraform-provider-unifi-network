# Basic firewall policy - allow traffic
resource "unifi_firewall_policy" "allow_internal" {
  site_id     = "default"
  name        = "Allow Internal Traffic"
  description = "Allow all traffic within internal zone"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }

  destination_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }
}

# Block traffic between zones
resource "unifi_firewall_policy" "block_guest_to_internal" {
  site_id     = "default"
  name        = "Block Guest to Internal"
  description = "Block guest network from accessing internal resources"
  enabled     = true
  action      = "block"

  source_endpoint = {
    zone_id = unifi_firewall_zone.guest.id
  }

  destination_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }
}

# Allow specific ports
resource "unifi_firewall_policy" "allow_web" {
  site_id     = "default"
  name        = "Allow Web Traffic"
  description = "Allow HTTP and HTTPS traffic"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.guest.id
  }

  destination_endpoint = {
    zone_id     = unifi_firewall_zone.dmz.id
    port_ranges = ["80", "443"]
  }

  ip_protocol_scope = {
    protocol = "tcp"
  }
}

# Policy with IP address filter
resource "unifi_firewall_policy" "allow_dns_server" {
  site_id     = "default"
  name        = "Allow DNS Server"
  description = "Allow access to DNS server"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }

  destination_endpoint = {
    ip_addresses = ["8.8.8.8", "8.8.4.4"]
    port_ranges  = ["53"]
  }

  ip_protocol_scope = {
    protocol = "udp"
  }
}

# Policy with network filter
resource "unifi_firewall_policy" "allow_to_network" {
  site_id     = "default"
  name        = "Allow to Specific Network"
  description = "Allow traffic to specific network"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }

  destination_endpoint = {
    network_ids = [unifi_network.corporate.id]
  }
}

# Policy with schedule
resource "unifi_firewall_policy" "business_hours_only" {
  site_id     = "default"
  name        = "Business Hours Access"
  description = "Allow access only during business hours"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.guest.id
  }

  destination_endpoint = {
    zone_id = unifi_firewall_zone.dmz.id
  }

  schedule = {
    mode             = "custom"
    repeat_on_days   = ["monday", "tuesday", "wednesday", "thursday", "friday"]
    time_range_start = "09:00"
    time_range_end   = "18:00"
  }
}

# Policy with connection state filter
resource "unifi_firewall_policy" "established_only" {
  site_id     = "default"
  name        = "Allow Established"
  description = "Allow only established connections"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.dmz.id
  }

  destination_endpoint = {
    zone_id = unifi_firewall_zone.internal.id
  }

  connection_state_filter = {
    established = true
    related     = true
    new         = false
    invalid     = false
  }
}

# Block IoT from management
resource "unifi_firewall_policy" "block_iot_management" {
  site_id     = "default"
  name        = "Block IoT to Management"
  description = "Prevent IoT devices from accessing management network"
  enabled     = true
  action      = "block"

  source_endpoint = {
    zone_id = unifi_firewall_zone.iot.id
  }

  destination_endpoint = {
    zone_id = unifi_firewall_zone.management.id
  }
}

# Allow IoT to internet only
resource "unifi_firewall_policy" "iot_internet" {
  site_id     = "default"
  name        = "IoT Internet Access"
  description = "Allow IoT devices to access internet"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id = unifi_firewall_zone.iot.id
  }

  destination_endpoint = {
    matching_target = "internet"
  }
}

# Complex policy with all options
resource "unifi_firewall_policy" "complex" {
  site_id     = "default"
  name        = "Complex Policy"
  description = "Full-featured firewall policy example"
  enabled     = true
  action      = "allow"

  source_endpoint = {
    zone_id      = unifi_firewall_zone.internal.id
    network_ids  = [unifi_network.corporate.id]
    ip_addresses = ["10.0.0.0/24"]
    port_ranges  = []
    mac_addresses = []
  }

  destination_endpoint = {
    zone_id      = unifi_firewall_zone.dmz.id
    ip_addresses = ["192.168.100.0/24"]
    port_ranges  = ["80", "443", "8080-8090"]
  }

  ip_protocol_scope = {
    protocol = "tcp_udp"
  }

  schedule = {
    mode             = "custom"
    repeat_on_days   = ["monday", "tuesday", "wednesday", "thursday", "friday"]
    time_range_start = "08:00"
    time_range_end   = "20:00"
  }

  connection_state_filter = {
    established = true
    related     = true
    new         = true
    invalid     = false
  }
}
