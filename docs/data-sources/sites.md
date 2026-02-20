---
page_title: "unifi_sites Data Source - UniFi Network"
subcategory: ""
description: |-
  Fetches the list of UniFi sites.
---

# unifi_sites (Data Source)

Fetches the list of UniFi sites available to the authenticated user.

## Example Usage

```terraform
data "unifi_sites" "all" {}

output "site_names" {
  value = [for site in data.unifi_sites.all.sites : site.name]
}
```

## Schema

### Read-Only

- `sites` (List of Object) - List of UniFi sites.
  - `id` (String) - The unique identifier of the site.
  - `name` (String) - The name of the site.
  - `internal_reference` (String) - The internal reference of the site.
