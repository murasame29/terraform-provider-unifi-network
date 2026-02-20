# Configure the UniFi Network Provider
provider "unifi" {
  # API key can be set via UNIFI_API_KEY environment variable
  api_key = var.unifi_api_key

  # Optional: Override the base URL (defaults to https://api.ui.com)
  # base_url = "https://api.ui.com"
}

variable "unifi_api_key" {
  description = "UniFi Cloud API key"
  type        = string
  sensitive   = true
}
