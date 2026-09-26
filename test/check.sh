#!/usr/bin/env bash
set -euo pipefail

out=${1:?usage: check.sh <rendered repository>}
out=$(cd "$out" && pwd)
here=$(cd "$(dirname "$0")" && pwd)
home=$(mktemp -d)
trap 'rm -rf "$home"' EXIT

step() { printf '\n== %s\n' "$*"; }

step shellcheck
shellcheck -S warning "$out"/bin/* "$out/infra/opentofu/tofu.sh" "$out/policy/check.sh"

step "ovhctl builds and passes its tests"
(cd "$out/ops" && go vet ./... && go test -count=1 ./...)

step "policy tests"
conftest verify --policy "$out/policy" --data "$out/policy/data"

step "secrets are generated and encrypted"
export ANSIBLE_VAULT_PASSWORD_FILE=$home/vault-pass
"$out/bin/generate-secrets"
if "$out/bin/generate-secrets" 2>/dev/null; then
  echo "a second run overwrote the secrets" >&2
  exit 1
fi
for f in ansible/inventory/group_vars/stack/vault.yml ansible/inventory/host_vars/vps/vault.yml; do
  head -1 "$out/$f" | grep -q '^\$ANSIBLE_VAULT' || { echo "$f is not encrypted" >&2; exit 1; }
done
openssl x509 -in "$out/ansible/files/ca.pem" -noout -ext basicConstraints,keyUsage | grep -q 'CA:TRUE'
openssl x509 -in "$out/ansible/files/ca.pem" -noout -ext keyUsage | grep -q 'Certificate Sign'

step "opentofu validates"
(cd "$out/infra/opentofu" && tofu fmt -check && TF_VAR_state_passphrase=validate-only-passphrase tofu init -backend=false -input=false >/dev/null && TF_VAR_state_passphrase=validate-only-passphrase tofu validate)

step "every variable the roles read resolves"
(cd "$out/ansible" && ansible-galaxy collection install -r requirements.yml -p .collections >/dev/null)
(cd "$out/ansible" && STACK_HOST=192.0.2.10 ansible-playbook "$here/resolve.yml" --check --limit vps)

step "playbook syntax"
(cd "$out/ansible" && STACK_HOST=192.0.2.10 ansible-playbook playbooks/provision.yml --syntax-check)

step "the Vault init output moves into the vault file"
mkdir -p "$out/secrets/vault"
printf '{"keys":["a","b","c","d","e"],"keys_base64":["k1","k2","k3","k4","k5"],"root_token":"hvs.test-root"}\n' >"$out/secrets/vault/init.json"
"$out/bin/import-vault-init"
[[ ! -e $out/secrets/vault/init.json ]] || { echo "init.json was left behind" >&2; exit 1; }
(cd "$out/ansible" && STACK_HOST=192.0.2.10 ansible-playbook "$here/resolve.yml" --check --limit vps \
  -e '{"check_vault_import": true, "expected_unseal_keys": ["k1", "k2", "k3", "k4", "k5"], "expected_root_token": "hvs.test-root"}')

printf '\nall checks passed for %s\n' "$out"
