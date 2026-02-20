# Basic WiFi broadcast with WPA2
resource "unifi_wifi_broadcast" "basic" {
  site_id       = "default"
  name          = "Basic WiFi"
  enabled       = true
  network_id    = unifi_network.basic.id
  security_type = "wpa2"
  passphrase    = "secure-password-here"
}

# Guest WiFi with client isolation
resource "unifi_wifi_broadcast" "guest" {
  site_id       = "default"
  name          = "Guest WiFi"
  enabled       = true
  network_id    = unifi_network.guest.id
  security_type = "wpa2"
  passphrase    = "guest-password"

  hide_name                = false
  client_isolation_enabled = true
}

# Hidden WiFi network
resource "unifi_wifi_broadcast" "hidden" {
  site_id       = "default"
  name          = "Hidden Network"
  enabled       = true
  network_id    = unifi_network.basic.id
  security_type = "wpa3"
  passphrase    = "very-secure-password"

  hide_name = true
}

# WiFi with WPA3 security
resource "unifi_wifi_broadcast" "wpa3" {
  site_id       = "default"
  name          = "WPA3 Network"
  enabled       = true
  network_id    = unifi_network.corporate.id
  security_type = "wpa3"
  passphrase    = "wpa3-secure-password"

  security_configuration = {
    pmf_mode                    = "required"
    group_rekey_interval        = 3600
    sae_anti_clogging_threshold = 5
    sae_sync_limit              = 5
  }
}

# WiFi with enterprise authentication (RADIUS)
resource "unifi_wifi_broadcast" "enterprise" {
  site_id       = "default"
  name          = "Enterprise WiFi"
  enabled       = true
  network_id    = unifi_network.corporate.id
  security_type = "wpa2-enterprise"

  security_configuration = {
    radius_profile_id = "radius-profile-id"
    pmf_mode          = "optional"
  }
}

# WiFi with device filter (specific APs only)
resource "unifi_wifi_broadcast" "filtered" {
  site_id       = "default"
  name          = "Filtered WiFi"
  enabled       = true
  network_id    = unifi_network.basic.id
  security_type = "wpa2"
  passphrase    = "filtered-password"

  broadcasting_device_filter = {
    mode       = "include"
    device_ids = ["ap-device-id-1", "ap-device-id-2"]
  }
}

# WiFi with specific frequency configuration
resource "unifi_wifi_broadcast" "dual_band" {
  site_id       = "default"
  name          = "Dual Band WiFi"
  enabled       = true
  network_id    = unifi_network.basic.id
  security_type = "wpa2"
  passphrase    = "dual-band-password"

  frequencies = {
    band_2_4_ghz = {
      enabled           = true
      min_rate_mbps     = 12
      multicast_enhance = true
    }
    band_5_ghz = {
      enabled           = true
      min_rate_mbps     = 24
      multicast_enhance = true
    }
    band_6_ghz = {
      enabled           = false
      min_rate_mbps     = 0
      multicast_enhance = false
    }
  }
}

# WiFi with band steering
resource "unifi_wifi_broadcast" "band_steering" {
  site_id       = "default"
  name          = "Band Steering WiFi"
  enabled       = true
  network_id    = unifi_network.basic.id
  security_type = "wpa2"
  passphrase    = "band-steering-password"

  band_steering = {
    mode = "prefer_5ghz"
  }
}

# WiFi with MLO (Multi-Link Operation) for WiFi 7
resource "unifi_wifi_broadcast" "wifi7_mlo" {
  site_id       = "default"
  name          = "WiFi 7 MLO"
  enabled       = true
  network_id    = unifi_network.basic.id
  security_type = "wpa3"
  passphrase    = "wifi7-mlo-password"

  mlo = {
    enabled = true
  }
}

# Full-featured corporate WiFi
resource "unifi_wifi_broadcast" "corporate" {
  site_id       = "default"
  name          = "Corporate WiFi"
  enabled       = true
  network_id    = unifi_network.corporate.id
  security_type = "wpa3"
  passphrase    = "corporate-secure-password"

  hide_name                = false
  client_isolation_enabled = false

  security_configuration = {
    pmf_mode             = "required"
    group_rekey_interval = 3600
  }

  frequencies = {
    band_2_4_ghz = {
      enabled           = true
      min_rate_mbps     = 12
      multicast_enhance = true
    }
    band_5_ghz = {
      enabled           = true
      min_rate_mbps     = 24
      multicast_enhance = true
    }
    band_6_ghz = {
      enabled           = true
      min_rate_mbps     = 24
      multicast_enhance = true
    }
  }

  band_steering = {
    mode = "prefer_5ghz"
  }

  broadcasting_device_filter = {
    mode       = "all"
    device_ids = []
  }
}
