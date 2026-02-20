# Get details of a specific network
data "unifi_network" "example" {
  site_id = "your-site-id"
  id      = "your-network-id"
}

output "network_name" {
  value = data.unifi_network.example.name
}
