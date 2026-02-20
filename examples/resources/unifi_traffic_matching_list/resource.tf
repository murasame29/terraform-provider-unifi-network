# Traffic matching list with port items
resource "unifi_traffic_matching_list" "web_ports" {
  site_id     = "default"
  name        = "Web Ports"
  description = "Common web server ports"

  port_items = [
    {
      protocol   = "tcp"
      port_range = "80"
    },
    {
      protocol   = "tcp"
      port_range = "443"
    },
    {
      protocol   = "tcp"
      port_range = "8080"
    }
  ]
}

# Traffic matching list with port ranges
resource "unifi_traffic_matching_list" "high_ports" {
  site_id     = "default"
  name        = "High Ports"
  description = "High port range for dynamic connections"

  port_items = [
    {
      protocol   = "tcp"
      port_range = "49152-65535"
    },
    {
      protocol   = "udp"
      port_range = "49152-65535"
    }
  ]
}

# Traffic matching list with IPv4 addresses
resource "unifi_traffic_matching_list" "internal_servers" {
  site_id     = "default"
  name        = "Internal Servers"
  description = "Internal server IP addresses"

  ip_address_items = [
    {
      address = "192.168.1.10"
    },
    {
      address = "192.168.1.20"
    },
    {
      address = "192.168.1.30"
    }
  ]
}

# Traffic matching list with IPv4 subnets
resource "unifi_traffic_matching_list" "internal_subnets" {
  site_id     = "default"
  name        = "Internal Subnets"
  description = "Internal network subnets"

  ip_address_items = [
    {
      address = "10.0.0.0/8"
    },
    {
      address = "172.16.0.0/12"
    },
    {
      address = "192.168.0.0/16"
    }
  ]
}

# Traffic matching list with IPv6 addresses
resource "unifi_traffic_matching_list" "ipv6_servers" {
  site_id     = "default"
  name        = "IPv6 Servers"
  description = "IPv6 server addresses"

  ipv6_address_items = [
    {
      address = "2001:db8::10"
    },
    {
      address = "2001:db8::20"
    }
  ]
}

# Traffic matching list with IPv6 subnets
resource "unifi_traffic_matching_list" "ipv6_subnets" {
  site_id     = "default"
  name        = "IPv6 Subnets"
  description = "IPv6 network subnets"

  ipv6_address_items = [
    {
      address = "2001:db8::/32"
    },
    {
      address = "fd00::/8"
    }
  ]
}

# Mixed traffic matching list
resource "unifi_traffic_matching_list" "database_access" {
  site_id     = "default"
  name        = "Database Access"
  description = "Database servers and ports"

  port_items = [
    {
      protocol   = "tcp"
      port_range = "3306"
    },
    {
      protocol   = "tcp"
      port_range = "5432"
    },
    {
      protocol   = "tcp"
      port_range = "27017"
    }
  ]

  ip_address_items = [
    {
      address = "192.168.10.100"
    },
    {
      address = "192.168.10.101"
    }
  ]
}

# Gaming ports
resource "unifi_traffic_matching_list" "gaming_ports" {
  site_id     = "default"
  name        = "Gaming Ports"
  description = "Common gaming platform ports"

  port_items = [
    {
      protocol   = "tcp"
      port_range = "3074"
    },
    {
      protocol   = "udp"
      port_range = "3074"
    },
    {
      protocol   = "tcp"
      port_range = "3478-3480"
    },
    {
      protocol   = "udp"
      port_range = "3478-3480"
    }
  ]
}

# VoIP ports
resource "unifi_traffic_matching_list" "voip_ports" {
  site_id     = "default"
  name        = "VoIP Ports"
  description = "Voice over IP ports"

  port_items = [
    {
      protocol   = "udp"
      port_range = "5060-5061"
    },
    {
      protocol   = "udp"
      port_range = "10000-20000"
    }
  ]
}

# Blocked IP addresses
resource "unifi_traffic_matching_list" "blocked_ips" {
  site_id     = "default"
  name        = "Blocked IPs"
  description = "IP addresses to block"

  ip_address_items = [
    {
      address = "203.0.113.0/24"
    },
    {
      address = "198.51.100.0/24"
    }
  ]

  ipv6_address_items = [
    {
      address = "2001:db8:bad::/48"
    }
  ]
}
