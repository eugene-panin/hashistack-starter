# {{ .ProjectName }}

Your {{ if eq .Provider "ovh" }}OVH VPS{{ else }}server{{ end }}, set up as a small private cloud{{ if .MailEnabled }} and the mail server for {{ join ", " .MailDomains }}{{ end }}.
Everything about it is in this folder. Keep the folder in git: the server can
be rebuilt from it at any time.

Generated from [hashistack-starter](https://github.com/eugene-panin/hashistack-starter).

## Before you start

- **Run every command below in the starter's terminal.** It has all the
  tools. To open it, go to this folder in a terminal on your laptop and run:

  ```bash
  docker run -it --rm -v "$PWD:/work" -v "$HOME/.ssh:/root/.ssh:ro" \
    -v "$HOME/.config:/root/.config" ghcr.io/eugene-panin/hashistack-starter
  ```

- **Edit files with any editor on your laptop.** The terminal and your laptop
  see the same folder.
- **Any step can be run again.** Running a step twice is safe. If a step
  fails, fix what the error says and run the same step again.

## Setup

### 1. Create the keys the setup needs

Run:

```bash
make setup
```

{{ if eq .Provider "ovh" -}}
It creates `secrets/ovh-api.env` and `secrets/cloudflare.env`. They hold keys
to your accounts and never go into git, so keep a copy somewhere safe, such
as your password manager. Open them and fill in the values.
{{- else -}}
It creates `secrets/cloudflare.env`. It holds keys to your accounts and never
goes into git, so keep a copy somewhere safe, such as your password manager.
Open it and fill in the values.
{{- end }}
{{- if eq .Provider "ovh" }}

**OVH key.** The setup uses it to reinstall your VPS.

1. Open [this link](https://api.ovh.com/createToken/index.cgi?GET=/vps&GET=/vps/*/ips&GET=/vps/*/images/available&GET=/vps/*/images/available/*&GET=/vps/*/tasks&GET=/vps/*/tasks/*&POST=/vps/*/rebuild&POST=/vps/*/getConsoleUrl&POST=/vps/*/reboot&POST=/vps/*/start&POST=/vps/*/stop).
   It asks OVH for exactly the rights the setup needs.
2. Log in with your OVH account.
3. Name the key `{{ .ProjectName }}` and set its validity to "Unlimited".
4. Copy the three values OVH shows into `secrets/ovh-api.env`:
   `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET` and `OVH_CONSUMER_KEY`.
{{- end }}

**Cloudflare keys.** The setup uses them to publish the DNS names of your
server and to get its certificates. You need two keys, both made the same
way:

1. In [Cloudflare](https://dash.cloudflare.com), open "My Profile", then "API
   Tokens", then "Create Token".
2. Choose the "Edit zone DNS" template.
3. Under "Permissions", add a second line: "Zone", "Zone", "Read".
4. Under "Zone Resources", choose "Specific zone" and the domain, as listed
   below. Add more lines for more domains.
5. "Continue to summary", then "Create Token", and copy the token.

| Key | Domains | Goes into |
|---|---|---|
| names | `{{ .InfraDomain }}`{{ if .MailEnabled }}{{ range .MailDomains }}{{ if and (ne . $.InfraDomain) (not (has . $.ManualDnsDomains)) }}, `{{ . }}`{{ end }}{{ end }}{{ end }} | `CLOUDFLARE_API_TOKEN` |
| certificates | `{{ .InfraDomain }}` | `TRAEFIK_CLOUDFLARE_TOKEN` |

### 2. Generate your passwords

```bash
make secrets
```

It generates every password and key the server needs and stores them
encrypted in this folder, so they can safely go into git. The one password
that opens them is written to `~/.config/{{ .ProjectName }}/vault-pass` on
your laptop. **Save that file in your password manager.** Without it
nothing in this folder can be opened again.

{{ if eq .Provider "ovh" -}}
### 3. Reinstall the VPS

```bash
make vps-rebuild
make vps-rebuild CONFIRM=1
```

The first command shows what would happen. The second one erases the VPS and
installs `{{ .VpsImage }}` on it, with a user `{{ .OpsUser }}` that only your
SSH key can log in as. Anything on the VPS is lost.
{{- else -}}
### 3. Lock the server down

```bash
make bootstrap
```

If you log in to the server with a password, run
`make bootstrap ASK_PASS=1` instead and type the password when asked.

The command logs in to `{{ .ServerIp }}` as `{{ .BootstrapUser }}`. It makes a
user `{{ .OpsUser }}` that logs in with your SSH key, checks that this login
works, and only then turns off logins as root and with passwords. From now
on only your laptop gets in.
{{- end }}

### 4. Install the platform

```bash
make provision-deps
make provision
make vault-import-init
```

`make provision` takes about ten minutes and installs everything on the
server:

- the private network (WireGuard) and the firewall;
- the scheduler that runs your apps (Nomad);
- the service directory (Consul);
- the password store (Vault).

Its first run creates the keys of the password store. `make
vault-import-init` moves them into your encrypted passwords.

### 5. Connect your laptop to the private network

1. Open the WireGuard app and choose "Import tunnel from file".
2. Pick `ansible/clients/{{ index .WireguardClients 0 }}.conf` from this folder.
3. Turn the tunnel on.
{{- if gt (len .WireguardClients) 1 }}

The other devices get their files next to it: {{ range $i, $c := .WireguardClients }}{{ if $i }}{{ if gt $i 1 }}, {{ end }}`{{ $c }}.conf`{{ end }}{{ end }}.
{{- end }}

The admin pages only open while the tunnel is on.

### 6. Start the services
{{- if .MailEnabled }}

First give the mail server its name. The mail server needs it to get its
certificate:

1. In [Cloudflare](https://dash.cloudflare.com), open the domain of
   `{{ .MailHostname }}`, then "DNS", then "Add record".
2. Choose type "A", name `{{ .MailHostname }}`, and as the address
   {{ if eq .Provider "ovh" }}the one `make vps-host` prints{{ else }}`{{ .ServerIp }}`{{ end }}.
3. Turn "Proxy status" off, so that it shows "DNS only".
4. Save.

Then:
{{- end }}

```bash
make infra-init
make infra-plan
make infra-apply CONFIRM=1
```

`make infra-plan` lists what will be created: DNS names, certificates, the
web front door (Traefik){{ if .MailEnabled }}, the mail server{{ end }}. `make infra-apply CONFIRM=1`
creates it.

### 7. Check

With the tunnel on, open these pages:

- `https://nomad.{{ .InfraDomain }}`: your apps;
- `https://consul.{{ .InfraDomain }}`: your services;
- `https://vault.{{ .InfraDomain }}`: your passwords.
{{- if .MailEnabled }}

### 8. Mail

- In {{ if eq .Provider "ovh" }}the OVH manager{{ else }}your server provider's panel{{ end }}, set the "reverse DNS" of the
  server's address to `{{ .MailHostname }}`. Without it, other mail servers
  treat your mail as spam.
{{- if .ManualDnsDomains }}
- The DNS of {{ join ", " .ManualDnsDomains }} is not on Cloudflare. Run
  `make infra-output OUTPUT=mail_manual_records` and create the records it
  lists at your DNS provider.
{{- end }}
- Get the mailbox passwords with
  `make infra-output OUTPUT='-json mail_passwords'`.
- In your mail app:
  - server: `{{ .MailHostname }}`;
  - incoming: IMAP, port 993;
  - outgoing: SMTP, port 465;
  - user name: the full address, such as `{{ index .Mailboxes 0 }}@{{ index .MailDomains 0 }}`.
{{- end }}

### Save your work

Put this folder into a private git repository, for example on GitHub. The
keys that must not go there are already excluded:

- `secrets/`;
- `~/.config/{{ .ProjectName }}/vault-pass`, which lives outside this folder.

## Later

- **Change something in the services:** edit the files in `infra/opentofu`,
  then run `make infra-plan` and `make infra-apply CONFIRM=1`.
- **Update the server:** run `make provision`. `make provision-check` shows
  what it would change without changing anything.
- **Change a stored password:** `make vault-edit`.
{{- if eq .Provider "ovh" }}
- **Locked out of the VPS:** `make vps-console` opens its screen in the
  browser.
{{- end }}

## What is inside

| Part | Tool | What it does |
|---|---|---|
{{- if eq .Provider "ovh" }}
| Machine | `ovhctl` (Go, `ops/`) | Reinstalls the VPS with an SSH-only first login and a host key known in advance |
{{- else }}
| Machine | Ansible, `playbooks/bootstrap.yml` | Creates the ops user and closes every other SSH login |
{{- end }}
| Host | Ansible, [`eugene_panin.base`](https://github.com/eugene-panin/ansible-collection-base) and [`eugene_panin.hashistack`](https://github.com/eugene-panin/ansible-collection-hashistack) | WireGuard, firewall, Consul, Vault, Docker, Nomad |
| Services | OpenTofu, [`eugene-panin/hashistack/nomad`](https://search.opentofu.org/module/eugene-panin/hashistack/nomad) | Workload identity, Traefik{{ if .MailEnabled }}, mail{{ end }}, DNS records in Cloudflare |

Every plan passes the Conftest policies in `policy/` before it can be applied.
