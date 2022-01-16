---
layout: "yandex"
page_title: "Yandex: yandex_vpc_security_group_rule"
sidebar_current: "docs-yandex-datasource-vpc-security-group-rule"
description: |-
  Get information about a Yandex VPC Security Group Rule.
---

# yandex\_vpc\_security\_group\_rule

Get information about a Yandex VPC Security Group Rule. For more information, see
[Yandex.Cloud VPC](https://cloud.yandex.com/docs/vpc/concepts/security-groups).

```hcl
data "yandex_vpc_security_group_rule" "rule1" {
  security_group_binding = "my-sg-id"
  rule_id                = "my-rule-id"
}
```

This data source used to define Security Group Rule that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `security_group_binding` (Required) - ID of Security Group that owns the rule.
* `rule_id` - (Required) - ID of the Security Group Rule.

## Attributes Reference

The following attribute is exported:
* `description` - Description of the rule.
* `direction` - Direction of the rule. Can be `ingress` (inbound) or `egress` (outbound).
* `labels` - Labels to assign to this rule.
* `protocol` - One of `ANY`, `TCP`, `UDP`, `ICMP`, `IPV6_ICMP`.
* `from_port` - Minimum port number.
* `to_port` - Maximum port number.
* `port` - Port number (if applied to a single port).
* `security_group_id` - Target security group ID for this rule.
* `predefined_target` - Special-purpose targets. `self_security_group` refers to this particular security group. `loadbalancer_healthchecks` represents [loadbalancer health check nodes](https://cloud.yandex.com/docs/network-load-balancer/concepts/health-check).
* `v4_cidr_blocks` - The blocks of  IPv4 addresses for this rule.
* `v6_cidr_blocks` - The blocks of  IPv6 addresses for this rule.
