#!/usr/bin/env bash
set -euo pipefail

policy=$(cd "$(dirname "$0")" && pwd)
plan=${1:?usage: check.sh <plan file relative to infra/opentofu, or a plan in JSON>}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

if [[ $plan == *.json ]]; then
  cp "$plan" "$work/plan.json"
else
  "$policy/../infra/opentofu/tofu.sh" show -json "$plan" >"$work/plan.json"
fi

conftest test --no-color --policy "$policy" --data "$policy/data" --namespace terraform "$work/plan.json"

mkdir "$work/jobs"
jq -r '.resource_changes[]
  | select(.type == "nomad_job" and .change.after.jobspec != null)
  | "\(.address)\t\(.change.after.jobspec | @base64)"' "$work/plan.json" |
  while IFS=$'\t' read -r address spec; do
    base64 -d <<<"$spec" >"$work/jobs/$address.nomad.hcl"
  done

if compgen -G "$work/jobs/*.nomad.hcl" >/dev/null; then
  conftest test --no-color --parser hcl2 --policy "$policy" --data "$policy/data" --namespace nomad "$work"/jobs/*.nomad.hcl
fi
