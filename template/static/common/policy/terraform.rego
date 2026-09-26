package terraform

import data.stack

infra_name(name) if name == stack.infra_zone

infra_name(name) if endswith(name, concat("", [".", stack.infra_zone]))

records contains [rc.address, rc.change.after] if {
	some rc in input.resource_changes
	rc.type == "cloudflare_dns_record"
	rc.change.after != null
	infra_name(rc.change.after.name)
}

deny contains msg if {
	some [address, record] in records
	record.proxied == true
	msg := sprintf("%s: %s is an internal name and must not go through the Cloudflare proxy", [address, record.name])
}

deny contains msg if {
	some [address, record] in records
	record.type == "A"
	not net.cidr_contains(stack.infra_cidr, record.content)
	msg := sprintf("%s: %s must point into %s, not %s", [address, record.name, stack.infra_cidr, record.content])
}

deny contains msg if {
	some [address, record] in records
	record.type == "AAAA"
	msg := sprintf("%s: %s is an internal name and has no IPv6 address inside WireGuard", [address, record.name])
}

deny contains msg if {
	some rc in input.resource_changes
	rc.type in stack.protected_types
	"delete" in rc.change.actions
	not rc.address in stack.allow_destroy
	msg := sprintf("%s: would be destroyed; add the address to allow_destroy in policy/data/stack.yaml to allow it", [rc.address])
}

deny contains msg if {
	some rc in input.resource_changes
	rc.type == "vault_kv_secret_v2"
	rc.change.after.data_json != null
	msg := sprintf("%s: data_json keeps the secret in the state; use data_json_wo", [rc.address])
}

deny contains msg if {
	some rc in input.resource_changes
	rc.type == "vault_generic_secret"
	rc.change.after != null
	msg := sprintf("%s: vault_generic_secret keeps the secret in the state; use vault_kv_secret_v2 with data_json_wo", [rc.address])
}
