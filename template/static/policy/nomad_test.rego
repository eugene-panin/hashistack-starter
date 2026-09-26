package nomad_test

import data.nomad

job(name, task, network) := {"job": {name: [{"group": {"g": [{
	"network": network,
	"task": {"t": [task]},
}]}}]}}

good_task := {
	"driver": "docker",
	"config": [{"image": "mirror.gcr.io/library/nginx:1.29"}],
	"resources": [{"cpu": 100, "memory": 64}],
}

internal_port := [{"port": {"http": [{}]}}]

test_good_job_allowed if {
	count(nomad.deny) == 0 with input as job("web", good_task, internal_port)
}

test_raw_exec_denied if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"driver": "raw_exec", "config": [{"command": "/bin/sh"}]}), internal_port)
}

test_privileged_denied if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"config": [{"image": "nginx:1.29", "privileged": true}]}), internal_port)
}

test_extra_capability_denied if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"config": [{"image": "nginx:1.29", "cap_add": ["net_bind_service", "sys_admin"]}]}), internal_port)
}

test_allowed_capability if {
	count(nomad.deny) == 0 with input as job("web", object.union(good_task, {"config": [{"image": "nginx:1.29", "cap_add": ["NET_BIND_SERVICE"]}]}), internal_port)
}

test_untagged_image_denied if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"config": [{"image": "nginx"}]}), internal_port)
}

test_latest_image_denied if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"config": [{"image": "nginx:latest"}]}), internal_port)
}

test_registry_port_is_not_a_tag if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"config": [{"image": "registry.local:5000/nginx"}]}), internal_port)
}

test_digest_allowed if {
	count(nomad.deny) == 0 with input as job("web", object.union(good_task, {"config": [{"image": "nginx@sha256:0123456789abcdef"}]}), internal_port)
}

test_missing_memory_denied if {
	count(nomad.deny) == 1 with input as job("web", object.union(good_task, {"resources": [{"cpu": 100}]}), internal_port)
}

test_missing_resources_denied if {
	count(nomad.deny) == 1 with input as job("web", object.remove(good_task, ["resources"]), internal_port)
}

test_public_port_denied if {
	count(nomad.deny) == 1 with input as job("web", good_task, [{"port": {"https": [{"static": 443, "host_network": "public"}]}}])
}

test_public_port_allowed_for_traefik if {
	count(nomad.deny) == 0 with input as job("traefik", good_task, [{"port": {"https": [{"static": 443, "host_network": "public"}]}}])
}
