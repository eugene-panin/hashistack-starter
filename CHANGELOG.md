# Changelog

All notable changes are documented here.
This project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Boilerplate template for one OVH VPS with Consul, Vault and Nomad over
  WireGuard, Traefik, and optional mail, built from the ovh-stack-iac
  repository: Ansible inventory and playbook, the OpenTofu root module,
  Conftest policies, `ovhctl`, and `make secrets` / `make vault-import-init`
  for the secrets of a first run. CI renders two sets of answers and checks
  each result.
