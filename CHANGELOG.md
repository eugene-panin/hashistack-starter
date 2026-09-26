# Changelog

All notable changes are documented here.
This project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `Provider: ssh` for any Ubuntu server you can already log in to: `make
  bootstrap` creates the ops user with your key and then, logged in as that
  user, turns off root and password logins. No `ovhctl` or OVH credentials in
  that case. Tested against an Ubuntu 24.04 server with systemd in CI.
- An image, `ghcr.io/eugene-panin/hashistack-starter`, with Boilerplate,
  Ansible, OpenTofu, Conftest and Go: it generates the repository and runs its
  steps, so Docker is the only thing to install. CI runs every check inside it
  and publishes it for amd64 and arm64 on tags.

## [0.1.0] - 2026-09-26

### Added

- Boilerplate template for one OVH VPS with Consul, Vault and Nomad over
  WireGuard, Traefik, and optional mail, built from the ovh-stack-iac
  repository: Ansible inventory and playbook, the OpenTofu root module,
  Conftest policies, `ovhctl`, and scripts that generate the secrets of a
  first run and import the Vault init output. CI renders two sets of answers
  and checks each result.

### Fixed

- The Ansible password file was passed twice, so generating the secrets
  failed.
- The CA, WireGuard client and Vault init paths resolve from the inventory, not
  from the playbook.
