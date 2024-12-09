---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitbucket Provider"
subcategory: ""
description: |-
  
---

# bitbucket Provider





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `author` (Block, Optional) (see [below for nested schema](#nestedblock--author))
- `bitbucket` (Block, Optional) (see [below for nested schema](#nestedblock--bitbucket))

<a id="nestedblock--author"></a>
### Nested Schema for `author`

Required:

- `email` (String)
- `name` (String)


<a id="nestedblock--bitbucket"></a>
### Nested Schema for `bitbucket`

Required:

- `endpoint` (String)
- `username` (String)

Optional:

- `password` (String, Sensitive)
- `token` (String, Sensitive)