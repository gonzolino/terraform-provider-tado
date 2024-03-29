---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tado_geofencing Resource - terraform-provider-tado"
subcategory: ""
description: |-
  Controls geofencing of a home.
---

# tado_geofencing (Resource)

Controls geofencing of a home.

## Example Usage

```terraform
resource "tado_geofencing" "auto" {
  home_name = "My Home"
  presence  = "auto"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `home_name` (String) Name of the home this geofencing resource belongs to.
- `presence` (String) Whether somebody is present in the home. Can be one of 'auto', 'home' or 'away'.

### Read-Only

- `id` (String) ID of this geofencing resource. This should match the home_name.
