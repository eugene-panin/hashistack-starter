package terraform_test

import data.terraform

record(after) := {"resource_changes": [{
	"address": "cloudflare_dns_record.x",
	"type": "cloudflare_dns_record",
	"change": {"actions": ["create"], "after": after},
}]}

change(type, actions, after) := {"resource_changes": [{
	"address": sprintf("%s.x", [type]),
	"type": type,
	"change": {"actions": actions, "after": after},
}]}

denied(after) := n if {
	n := count(terraform.deny) with input as record(after)
		with data.stack.infra_zone as "internal.test"
		with data.stack.infra_cidr as "10.99.0.0/24"
}

test_internal_record_allowed if {
	denied({"name": "*.internal.test", "type": "A", "content": "10.99.0.1", "proxied": false}) == 0
}

test_proxied_internal_record_denied if {
	denied({"name": "*.internal.test", "type": "A", "content": "10.99.0.1", "proxied": true}) == 1
}

test_internal_record_outside_wireguard_denied if {
	denied({"name": "app.internal.test", "type": "A", "content": "192.0.2.10", "proxied": false}) == 1
}

test_internal_ipv6_record_denied if {
	denied({"name": "app.internal.test", "type": "AAAA", "content": "2001:db8::1", "proxied": false}) == 1
}

test_other_zone_not_checked if {
	denied({"name": "www.example.org", "type": "A", "content": "192.0.2.10", "proxied": true}) == 0
}

test_lookalike_zone_not_internal if {
	denied({"name": "notinternal.test", "type": "A", "content": "192.0.2.10", "proxied": true}) == 0
}

test_protected_destroy_denied if {
	count(terraform.deny) == 1 with input as change("vault_mount", ["delete"], null)
}

test_protected_replace_denied if {
	count(terraform.deny) == 1 with input as change("nomad_dynamic_host_volume", ["delete", "create"], {})
}

test_allowed_destroy if {
	count(terraform.deny) == 0 with input as change("vault_mount", ["delete"], null)
		with data.stack.allow_destroy as ["vault_mount.x"]
}

test_unprotected_destroy_allowed if {
	count(terraform.deny) == 0 with input as change("nomad_job", ["delete"], null)
}

test_write_only_secret_allowed if {
	count(terraform.deny) == 0 with input as change("vault_kv_secret_v2", ["create"], {"data_json": null, "data_json_wo": null})
}

test_secret_in_state_denied if {
	count(terraform.deny) == 1 with input as change("vault_kv_secret_v2", ["create"], {"data_json": "{\"k\":\"v\"}"})
}

test_generic_secret_denied if {
	count(terraform.deny) == 1 with input as change("vault_generic_secret", ["create"], {"data_json": "{}"})
}
