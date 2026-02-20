# Basic network with VLAN
resource "unifi_network" "basic" {
  site_id    = "default"
  name       = "Basic Network"
  enabled    = true
  vlan_id    = 10
  management = "third-party"
}

# Guest network with isolation
resource "unifi_network" "guest" {
  site_id    = "default"
  name       = "Guest Network"
  enabled    = true
  vlan_id    = 100
  management = "third-party"

  isolation_enabled       = true
  internet_access_enabled = true
  mdns_forwarding_enabled = false
}

# Network with IPv4 DHCP configuration
resource "unifi_network" "dhcp_network" {
  site_id    = "default"
  name       = "DHCP Network"
  enabled    = true
  vlan_id    = 20
  management = "third-party"

  ipv4_configuration = {
    static_subnet = "192.168.20.0/24"
    gateway       = "192.168.20.1"
    dns_servers   = ["8.8.8.8", "8.8.4.4"]
    dhcp = {
      mode               = "server"
      start              = "192.168.20.100"
      end                = "192.168.20.200"
      lease_time_seconds = 86400
      dns_servers        = ["192.168.20.1"]
      gateway            = "192.168.20.1"
      boot_enabled       = false
    }
  }
}

# Network with IPv6 configuration
resource "unifi_network" "ipv6_network" {
  site_id    = "default"
  name       = "IPv6 Network"
  enabled    = true
  vlan_id    = 30
  management = "third-party"

  ipv6_configuration = {
    mode          = "static"
    static_subnet = "2001:db8::/64"
    gateway       = "2001:db8::1"
    dns_servers   = ["2001:4860:4860::8888", "2001:4860:4860::8844"]
    ra = {
      enabled                    = true
      mode                       = "slaac"
      priority                   = "high"
      valid_lifetime_seconds     = 86400
      preferred_lifetime_seconds = 14400
    }
  }
}

# Network with NAT outbound configuration
resource "unifi_network" "nat_network" {
  site_id    = "default"
  name       = "NAT Network"
  enabled    = true
  vlan_id    = 40
  management = "third-party"

  nat_outbound = {
    mode = "auto"
  }
}

# Network with DHCP guarding
resource "unifi_network" "secure_network" {
  site_id    = "default"
  name       = "Secure Network"
  enabled    = true
  vlan_id    = 50
  management = "third-party"

  dhcp_guarding = {
    enabled                 = true
    allowed_dhcp_server_ids = ["00:11:22:33:44:55"]
  }
}

# Corporate network with full configuration
resource "unifi_network" "corporate" {
  site_id    = "default"
  name       = "Corporate Network"
  enabled    = true
  vlan_id    = 1
  management = "third-party"

  isolation_enabled       = false
  internet_access_enabled = true
  mdns_forwarding_enabled = true

  ipv4_configuration = {
    static_subnet = "10.0.0.0/24"
    gateway       = "10.0.0.1"
    dns_servers   = ["10.0.0.1"]
    dhcp = {
      mode               = "server"
      start              = "10.0.0.50"
      end                = "10.0.0.250"
      lease_time_seconds = 43200
      dns_servers        = ["10.0.0.1", "8.8.8.8"]
      gateway            = "10.0.0.1"
      boot_enabled       = false
      ntp_servers        = ["pool.ntp.org"]
      tftp_server        = ""
      wins_servers       = []
    }
  }

  nat_outbound = {
    mode = "auto"
  }
}
