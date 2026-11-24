---
page_title: "Upgrading to v2"
subcategory: ""
description: |-
  Changes to make to configurations written for v0.2.x of the provider.
---

# Upgrading from v0.2.x to v2

v2 moves the provider to the v2 Fivetran Activations API. Workspaces are gone, destinations are replaced by activations, and syncs belong to an activation instead of a workspace. None of the steps below create, replace or delete anything in Fivetran Activations.

## Before you start

* **Back up your state.** The first plan with v2 upgrades `census_source` and `census_sync` state, and v0.2.x can't read the upgraded state.
* **Don't remove `census_workspace` or `census_destination` blocks while still on v0.2.x.** On v0.2.x that deletes them, and their syncs, in Fivetran Activations. On v2 removing them makes no API calls.
* Make all the configuration changes below in one pass. v2 rejects the old arguments with `Unsupported argument`, so a partly migrated configuration won't plan.

## Summary

| Resource | v0.2.x | v2 |
|---|---|---|
| provider | `region` | Removed. Use `base_url` if you need a non-default host. |
| `census_workspace` | Manages a workspace | No-op. Remove it. |
| `census_destination` | Manages a destination | No-op. Replace with `census_activation`. |
| `census_source` | `workspace_id` | Removed |
| `census_dataset` | `workspace_id` | Removed |
| `census_dataset` | `source_id = census_source.x.id` | `group_id = census_source.x.group_id` |
| `census_sync` | `workspace_id` | `activation_id = census_activation.x.id` |
| `census_sync` | `source_attributes.connection_id` | Removed |
| `census_sync` | `destination_attributes.connection_id` | Removed |
| `census_sync` | `run_mode.type = "live"` | Rejected. Only `triggered` is accepted. |
| `census_sync` | `field_mapping` with no primary identifier | Rejected. One mapping needs `is_primary_identifier = true`. |
| `census_sync` import | `workspace_id:sync_id` | `sync_id` |
| `census_source` import | `workspace_id:source_id` | `source_id` |
| `census_dataset` import | `workspace_id:dataset_id` | `dataset_id` |

## 1. Upgrade the provider

```diff
 terraform {
   required_providers {
     census = {
       source  = "sutrolabs/census"
-      version = "~> 0.2.0"
+      version = "~> 2.0"
     }
   }
 }

 provider "census" {
   personal_access_token = var.census_personal_token
-  region                = "us"
 }
```

```bash
terraform init -upgrade
```

## 2. Remove workspaces

Delete every `census_workspace` resource and data source, and every `workspace_id` argument on `census_source`, `census_dataset` and the `census_sync` data source. On `census_sync` resources, `workspace_id` becomes `activation_id` (step 5).

```diff
-resource "census_workspace" "main" {
-  name = "Production"
-}
-
 resource "census_source" "warehouse" {
-  workspace_id = census_workspace.main.id
-  name         = "Production Warehouse"
-  type         = "snowflake"
+  name = "Production Warehouse"
+  type = "snowflake"

   connection_config = { ... }
 }
```

## 3. Replace destinations with activations

An activation is a (source, destination) pair. Fivetran Activations has already created one for every source and destination your syncs use, so import them instead of creating new ones.

Declare one `census_activation` per (source, destination) pair your syncs use. `credentials` takes the same keys as the destination's `connection_config`.

```diff
-resource "census_destination" "salesforce" {
-  workspace_id = census_workspace.main.id
-  name         = "Production Salesforce"
-  type         = "salesforce"
-
-  connection_config = {
-    instance_url = "https://mycompany.my.salesforce.com"
-    username     = var.salesforce_username
-    password     = var.salesforce_password
-  }
-}
+resource "census_activation" "salesforce" {
+  group_id = census_source.warehouse.group_id
+  name     = "Production Salesforce"
+  type     = "salesforce"
+
+  credentials = {
+    instance_url = "https://mycompany.my.salesforce.com"
+    username     = var.salesforce_username
+    password     = var.salesforce_password
+  }
+}
```

| `census_destination` | `census_activation` |
|---|---|
| `workspace_id` | `group_id` of the source the syncs read from |
| `name` | `name` |
| `type` | `type` |
| `connection_config` | `credentials` |
| `auto_refresh_objects` | Removed |

A destination used by syncs from two sources becomes two activations, one per source.

## 4. Point datasets at the source's group

```diff
 resource "census_dataset" "active_users" {
-  workspace_id = census_workspace.main.id
-  source_id    = census_source.warehouse.id
-  name         = "Active Users"
+  group_id = census_source.warehouse.group_id
+  name     = "Active Users"

   query = "select id, email from users where active"
 }
```

The dataset is not replaced.

## 5. Update syncs

```diff
 resource "census_sync" "users_to_salesforce" {
-  workspace_id = census_workspace.main.id
-  label        = "Users to Salesforce"
-  operation    = "upsert"
+  activation_id = census_activation.salesforce.id
+  label         = "Users to Salesforce"
+  operation     = "upsert"

   source_attributes {
-    connection_id = census_source.warehouse.id
     object {
       type          = "table"
       table_name    = "users"
       table_schema  = "public"
       table_catalog = "prod"
     }
   }

   destination_attributes {
-    connection_id = census_destination.salesforce.id
-    object        = "Contact"
+    object = "Contact"
   }

   field_mapping {
     from                  = "email"
     to                    = "Email"
     is_primary_identifier = true
   }
 }
```

The activation must be the one for the sync's existing source and destination. `activation_id` forces replacement, so a different activation shows up in the plan as a destroy and create.

## 6. Update data source references

| v0.2.x | v2 |
|---|---|
| `data.census_sync.x` with `workspace_id = ...` | Remove `workspace_id` |
| `data.census_sync.x.source_attributes[0].object` | `data.census_sync.x.source_object` |
| `data.census_sync.x.destination_attributes[0].object` | `data.census_sync.x.destination_object` |
| `data.census_sync.x.source_attributes[0].connection_id` | Removed |
| `data.census_sync.x.destination_attributes[0].connection_id` | Removed. Use `data.census_sync.x.activation_id`. |
| `data.census_sync.x.field_mapping` | Removed |
| `data.census_sync.x.last_run_at`, `next_run_at`, `last_run_id` | Removed |
| `data.census_source.x` with `workspace_id = ...` | Remove `workspace_id` |
| `data.census_dataset.x` with `workspace_id = ...` | Remove `workspace_id` |
| `data.census_dataset.x.source_id` | `data.census_dataset.x.group_id` |
| `data.census_destination.x` | `data.census_activation.x`, looked up by `id = census_activation.x.id`, or by `group_id` plus `destination_id` (the old destination ID). The destination data source returns only `id`. |
| `data.census_workspace.x` | Remove. Returns only `id`. |

## 7. Import the activations

Importing an activation for a v0.2.x destination takes the source's `group_id` and the destination's ID:

```bash
terraform import census_activation.salesforce <group_id>:<destination_id>
```

`destination_id` is the ID of the `census_destination` you removed. Look it up in your backed-up state, or with `terraform state show census_destination.salesforce` if you haven't run a plan yet.

`group_id` isn't in v0.2.x state. Get it from the v2 API:

```bash
curl -s -H "Authorization: Bearer $CENSUS_PERSONAL_ACCESS_TOKEN" \
  https://app.getcensus.com/api/v2/sources/<source_id> | jq -r '.data.group_id'
```

## 8. Plan

```bash
terraform plan
```

Expected changes:

* An in-place update to `credentials` on each imported `census_activation`. The API never returns credentials, so Terraform sends the configured ones on apply.
* Removal of any `census_workspace` and `census_destination` still in state. No API calls are made.

Anything else is a mistake in the migrated configuration. In particular, a `census_sync` or `census_dataset` marked for replacement means its `activation_id` or `group_id` doesn't match what it already uses.
