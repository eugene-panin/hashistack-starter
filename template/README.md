# {{ .ProjectName }}

One {{ if eq .Provider "ovh" }}OVH VPS{{ else }}server{{ end }} running Consul, Vault and Nomad, reachable over WireGuard, with
Traefik in front{{ if .MailEnabled }} and a mail server for {{ join ", " .MailDomains }}{{ end }}.
Generated from [hashistack-starter](https://github.com/eugene-panin/hashistack-starter).

| Layer | Tool | What it does |
|---|---|---|
{{- if eq .Provider "ovh" }}
| Machine | `ovhctl` (Go, `ops/`) | Rebuilds the VPS with an SSH-only bootstrap and a host key it knows in advance |
{{- else }}
| Machine | Ansible, `playbooks/bootstrap.yml` | Creates the ops user and closes every other SSH login |
{{- end }}
| Hosts | Ansible, [`eugene_panin.base`](https://github.com/eugene-panin/ansible-collection-base) and [`eugene_panin.hashistack`](https://github.com/eugene-panin/ansible-collection-hashistack) | WireGuard, firewall, Consul, Vault, Docker, Nomad |
| Resources | OpenTofu, [`eugene-panin/hashistack/nomad`](https://search.opentofu.org/module/eugene-panin/hashistack/nomad) | Workload identity, Traefik{{ if .MailEnabled }}, mail{{ end }}, DNS records in Cloudflare |

Internal names live under `{{ .InfraDomain }}` and answer only through
WireGuard: `https://consul.{{ .InfraDomain }}`, `https://nomad.{{ .InfraDomain }}`,
`https://vault.{{ .InfraDomain }}`.

## Tools

Every command below runs inside the starter's image, which holds all the tools;
you need Docker and WireGuard on your machine:

```bash
docker run -it --rm -v "$PWD:/work" -v "$HOME/.ssh:/root/.ssh:ro" \
  -v "$HOME/.config/{{ .ProjectName }}:/root/.config/{{ .ProjectName }}" \
  ghcr.io/eugene-panin/hashistack-starter
```

Or install them yourself: `ansible-core` with Python's PyYAML, OpenTofu >= 1.11,
`conftest`, `jq`, `openssl`{{ if eq .Provider "ovh" }}, Go{{ end }}.

## From zero

1. **Credentials.** `make setup` creates {{ if eq .Provider "ovh" }}`secrets/ovh-api.env` and
   {{ end }}`secrets/cloudflare.env` (gitignored). Fill in:
{{- if eq .Provider "ovh" }}
   - an OVH API token with the rights `ovhctl` uses:
     `https://api.ovh.com/createToken/index.cgi?GET=/vps&GET=/vps/*/ips&GET=/vps/*/images/available&GET=/vps/*/images/available/*&GET=/vps/*/tasks&GET=/vps/*/tasks/*&POST=/vps/*/rebuild&POST=/vps/*/getConsoleUrl&POST=/vps/*/reboot&POST=/vps/*/start&POST=/vps/*/stop`
{{- end }}
   - `CLOUDFLARE_API_TOKEN`: Zone DNS Edit and Zone Read on every zone OpenTofu
     manages records in;
   - `TRAEFIK_CLOUDFLARE_TOKEN`: Zone DNS Edit on `{{ .InfraDomain }}` only,
     for the wildcard certificate.

2. **Secrets.** `make secrets` generates the CA, the gossip keys, the ACL
   tokens and the state passphrase, and encrypts them into
   `ansible/inventory/**/vault.yml`. The vault password is written to
   `~/.config/{{ .ProjectName }}/vault-pass`: back it up, nothing can be
   decrypted without it. Commit the vault files and `ansible/files/ca.pem`.

{{- if eq .Provider "ovh" }}
3. **VPS.** `make vps-rebuild` shows what would be sent; `make vps-rebuild CONFIRM=1`
   wipes the VPS and installs `{{ .VpsImage }}` with the `{{ .OpsUser }}` user.
{{- else }}
3. **Server.** `make bootstrap` logs in to `{{ .ServerIp }}` as `{{ .BootstrapUser }}`
   (`make bootstrap ASK_PASS=1` if that login takes a password), creates the
   `{{ .OpsUser }}` user with `{{ .SshPublicKeyPath }}`, then, logged in as
   `{{ .OpsUser }}`, turns off root and password logins. The server is not
   reinstalled; it should be a fresh Ubuntu 24.04.
{{- end }}

4. **Hosts.** `make provision-deps`, then `make provision`. The first run
   initialises Vault and writes its unseal keys and root token to
   `secrets/vault/init.json`; `make vault-import-init` moves them into the vault
   file. Commit it. `make provision-check` on a converged host reports
   `changed=0`.

5. **WireGuard.** Import `ansible/clients/<name>.conf` into the WireGuard app of
   each client: {{ join ", " .WireguardClients }}.

6. **Resources.** `make infra-init`, `make infra-plan`, then
   `make infra-apply CONFIRM=1`. Every plan passes the Conftest policies in
   `policy/` before it can be applied.
{{- if .MailEnabled }}

7. **Mail.**
{{- if eq .Provider "ovh" }}
   - Point an A record for `{{ .MailHostname }}` at the VPS address
     (`make vps-host`), and set the same name as its reverse DNS in the OVH
     manager.
{{- else }}
   - Point an A record for `{{ .MailHostname }}` at `{{ .ServerIp }}`, and ask
     your provider to set the same name as the reverse DNS of that address.
{{- end }}
{{- if .ManualDnsDomains }}
   - Create the records of {{ join ", " .ManualDnsDomains }} by hand:
     `make infra-output OUTPUT=mail_manual_records`.
{{- end }}
   - Mailbox passwords: `make infra-output OUTPUT='-json mail_passwords'`.
     Clients connect to `{{ .MailHostname }}`, IMAP 993 and SMTP 465, with the
     full address as the user name.
{{- end }}

## Day to day

- `make infra-plan` and `make infra-apply CONFIRM=1` for anything under
  `infra/opentofu`; `make provision-check` and `make provision` for the host.
- `make vault-edit` to change a secret.
{{- if eq .Provider "ovh" }}
- `make vps-console` opens OVH's KVM console when SSH is locked out.
{{- end }}
