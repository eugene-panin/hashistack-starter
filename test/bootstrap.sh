#!/usr/bin/env bash
set -euo pipefail

out=${1:?usage: bootstrap.sh <rendered ssh repository> <private key> <port>}
key=${2:?}
port=${3:?}

ssh_as() {
  ssh -i "$key" -p "$port" -o BatchMode=yes -o StrictHostKeyChecking=accept-new \
    -o UserKnownHostsFile="$out/.known_hosts" "$1@127.0.0.1" "${@:2}"
}

bootstrap() {
  (cd "$out/ansible" && STACK_HOST=127.0.0.1 ansible-playbook playbooks/bootstrap.yml \
    -e ansible_port="$port" \
    -e ansible_ssh_private_key_file="$key" \
    -e ansible_ssh_common_args="-o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=$out/.known_hosts")
}

echo "== root can log in before the bootstrap"
ssh_as root true

echo "== bootstrap"
bootstrap

echo "== ops logs in with the key and uses sudo without a password"
ssh_as ops sudo -n true

echo "== root can no longer log in"
if ssh_as root true 2>/dev/null; then
  echo "root still logs in" >&2
  exit 1
fi

echo "== password logins are off"
ssh_as ops sudo sshd -T | grep -qx 'passwordauthentication no'
ssh_as ops sudo sshd -T | grep -qx 'permitrootlogin no'

echo "== a second run changes nothing"
bootstrap | tee "$out/.second-run"
grep -Eq 'changed=0 .*failed=0' "$out/.second-run"

echo "bootstrap checks passed"
