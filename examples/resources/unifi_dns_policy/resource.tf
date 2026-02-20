# Basic DNS policy with A record
resource "unifi_dns_policy" "a_record" {
  site_id     = "default"
  name        = "Internal Server"
  description = "A record for internal server"
  enabled     = true

  record = {
    type = "A"
    a = {
      hostname   = "server.local"
      ip_address = "192.168.1.100"
      ttl        = 3600
    }
  }
}

# DNS policy with AAAA record (IPv6)
resource "unifi_dns_policy" "aaaa_record" {
  site_id     = "default"
  name        = "IPv6 Server"
  description = "AAAA record for IPv6 server"
  enabled     = true

  record = {
    type = "AAAA"
    aaaa = {
      hostname     = "server6.local"
      ipv6_address = "2001:db8::100"
      ttl          = 3600
    }
  }
}

# DNS policy with CNAME record
resource "unifi_dns_policy" "cname_record" {
  site_id     = "default"
  name        = "Alias Record"
  description = "CNAME alias for server"
  enabled     = true

  record = {
    type = "CNAME"
    cname = {
      hostname = "www.local"
      target   = "server.local"
      ttl      = 3600
    }
  }
}

# DNS policy with MX record
resource "unifi_dns_policy" "mx_record" {
  site_id     = "default"
  name        = "Mail Server"
  description = "MX record for mail server"
  enabled     = true

  record = {
    type = "MX"
    mx = {
      hostname    = "example.local"
      mail_server = "mail.example.local"
      priority    = 10
      ttl         = 3600
    }
  }
}

# DNS policy with TXT record
resource "unifi_dns_policy" "txt_record" {
  site_id     = "default"
  name        = "SPF Record"
  description = "TXT record for SPF"
  enabled     = true

  record = {
    type = "TXT"
    txt = {
      hostname = "example.local"
      text     = "v=spf1 include:_spf.google.com ~all"
      ttl      = 3600
    }
  }
}

# DNS policy with SRV record
resource "unifi_dns_policy" "srv_record" {
  site_id     = "default"
  name        = "SIP Service"
  description = "SRV record for SIP service"
  enabled     = true

  record = {
    type = "SRV"
    srv = {
      service  = "_sip"
      protocol = "_tcp"
      hostname = "example.local"
      target   = "sipserver.example.local"
      port     = 5060
      priority = 10
      weight   = 100
      ttl      = 3600
    }
  }
}

# DNS policy with PTR record (reverse DNS)
resource "unifi_dns_policy" "ptr_record" {
  site_id     = "default"
  name        = "Reverse DNS"
  description = "PTR record for reverse lookup"
  enabled     = true

  record = {
    type = "PTR"
    ptr = {
      ip_address = "192.168.1.100"
      hostname   = "server.local"
      ttl        = 3600
    }
  }
}

# Multiple A records for load balancing
resource "unifi_dns_policy" "lb_server1" {
  site_id     = "default"
  name        = "LB Server 1"
  description = "Load balanced server 1"
  enabled     = true

  record = {
    type = "A"
    a = {
      hostname   = "app.local"
      ip_address = "192.168.1.101"
      ttl        = 300
    }
  }
}

resource "unifi_dns_policy" "lb_server2" {
  site_id     = "default"
  name        = "LB Server 2"
  description = "Load balanced server 2"
  enabled     = true

  record = {
    type = "A"
    a = {
      hostname   = "app.local"
      ip_address = "192.168.1.102"
      ttl        = 300
    }
  }
}

# LDAP SRV record for Active Directory
resource "unifi_dns_policy" "ldap_srv" {
  site_id     = "default"
  name        = "LDAP Service"
  description = "SRV record for LDAP"
  enabled     = true

  record = {
    type = "SRV"
    srv = {
      service  = "_ldap"
      protocol = "_tcp"
      hostname = "dc._msdcs.example.local"
      target   = "dc1.example.local"
      port     = 389
      priority = 0
      weight   = 100
      ttl      = 600
    }
  }
}
