# List all networks for a site
data "unifi_networks" "all" {
  site_id = "your-site-id"
}

output "networks" {
  value = data.unifi_networks.all.networks
}
