#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
root=$(cd ../.. && pwd)
ca=$root/ansible/files/ca.pem
address=${STACK_ADDRESS:?set STACK_ADDRESS, or run through make}
export ANSIBLE_VAULT_PASSWORD_FILE=${ANSIBLE_VAULT_PASSWORD_FILE:-$HOME/.config/$(basename "$root")/vault-pass}

vault=$(ansible-vault view "$root/ansible/inventory/group_vars/stack/vault.yml" </dev/null \
  | python3 -c 'import json, sys, yaml; print(json.dumps(yaml.safe_load(sys.stdin)))')

field() {
  local value
  value=$(jq -r --arg k "$1" '.[$k] // empty' <<<"$vault")
  [[ -n $value ]] || { echo "$1 is missing from the Ansible vault" >&2; exit 1; }
  printf '%s' "$value"
}

set -a
. "$root/secrets/cloudflare.env"
set +a
: "${CLOUDFLARE_API_TOKEN:?missing from secrets/cloudflare.env}"
: "${TRAEFIK_CLOUDFLARE_TOKEN:?missing from secrets/cloudflare.env}"

export TF_VAR_traefik_cloudflare_token=$TRAEFIK_CLOUDFLARE_TOKEN
unset TRAEFIK_CLOUDFLARE_TOKEN
TF_VAR_state_passphrase=$(field vault_tofu_state_passphrase)
export TF_VAR_state_passphrase

export CONSUL_HTTP_ADDR=https://$address:8501 CONSUL_CACERT=$ca
export VAULT_ADDR=https://$address:8200 VAULT_CACERT=$ca
export NOMAD_ADDR=https://$address:4646 NOMAD_CACERT=$ca
CONSUL_HTTP_TOKEN=$(field vault_consul_acl_bootstrap_token)
VAULT_TOKEN=$(field vault_root_token)
NOMAD_TOKEN=$(field vault_nomad_acl_bootstrap_token)
export CONSUL_HTTP_TOKEN VAULT_TOKEN NOMAD_TOKEN

exec tofu "$@"
