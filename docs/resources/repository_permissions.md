---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitbucket_repository_permissions Resource - bitbucket"
subcategory: ""
description: |-
  
---

# bitbucket_repository_permissions (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String)
- `slug` (String)

### Optional

- `assignment_version` (String)
- `assignments` (Block List) (see [below for nested schema](#nestedblock--assignments))
- `retain_on_delete` (Boolean)

### Read-Only

- `computed_groups` (Attributes List) (see [below for nested schema](#nestedatt--computed_groups))
- `computed_users` (Attributes List) (see [below for nested schema](#nestedatt--computed_users))

<a id="nestedblock--assignments"></a>
### Nested Schema for `assignments`

Required:

- `permission` (String)
- `priority` (Number)

Optional:

- `groups` (List of String)
- `users` (List of String)


<a id="nestedatt--computed_groups"></a>
### Nested Schema for `computed_groups`

Read-Only:

- `name` (String)
- `permission` (String)


<a id="nestedatt--computed_users"></a>
### Nested Schema for `computed_users`

Read-Only:

- `name` (String)
- `permission` (String)