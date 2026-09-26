package nomad

import data.stack

groups contains [job_name, group_name, group] if {
	some job_name, jobs in input.job
	some job in jobs
	some group_name, blocks in job.group
	some group in blocks
}

tasks contains [sprintf("%s/%s/%s", [job_name, group_name, task_name]), task] if {
	some [job_name, group_name, group] in groups
	some task_name, blocks in group.task
	some task in blocks
}

deny contains msg if {
	some [name, task] in tasks
	task.driver == "raw_exec"
	msg := sprintf("%s: raw_exec runs as root without isolation; use exec or docker", [name])
}

deny contains msg if {
	some [name, task] in tasks
	some config in task.config
	config.privileged == true
	msg := sprintf("%s: privileged containers are not allowed", [name])
}

deny contains msg if {
	some [name, task] in tasks
	some config in task.config
	some cap in config.cap_add
	not lower(cap) in stack.allowed_caps
	msg := sprintf("%s: capability %s is not in allowed_caps", [name, cap])
}

deny contains msg if {
	some [name, task] in tasks
	task.driver == "docker"
	some config in task.config
	not pinned(config.image)
	msg := sprintf("%s: image %s has no fixed tag or digest", [name, config.image])
}

deny contains msg if {
	some [name, task] in tasks
	not has_memory(task)
	msg := sprintf("%s: resources.memory is not set", [name])
}

deny contains msg if {
	some [job_name, group_name, group] in groups
	some network in group.network
	some port_name, ports in network.port
	some port in ports
	port.host_network == stack.public_host_network
	not job_name in stack.public_jobs
	msg := sprintf("%s/%s: port %s is on the public network; only %v may bind there, everything else goes through Traefik", [job_name, group_name, port_name, stack.public_jobs])
}

pinned(image) if contains(image, "@sha256:")

pinned(image) if {
	not contains(image, "@")
	parts := split(image, "/")
	last := parts[count(parts) - 1]
	contains(last, ":")
	not endswith(last, ":latest")
}

has_memory(task) if {
	some resources in task.resources
	resources.memory > 0
}
