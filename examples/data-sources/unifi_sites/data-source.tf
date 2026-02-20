# List all UniFi sites
data "unifi_sites" "all" {}

output "sites" {
  value = data.unifi_sites.all.sites
}
