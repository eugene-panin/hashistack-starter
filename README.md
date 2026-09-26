# hashistack-starter

A [Boilerplate](https://github.com/gruntwork-io/boilerplate) template that
generates a repository for one Ubuntu server running Consul, Vault and Nomad
over WireGuard, with Traefik in front and, if you want it, a mail server. The
server is either an OVH VPS, rebuilt through the OVH API, or any server you can
already log in to over SSH. The generated repository holds your answers and
your secrets; the logic lives in published pieces it pins:

- Ansible collections [`eugene_panin.base`](https://github.com/eugene-panin/ansible-collection-base)
  and [`eugene_panin.hashistack`](https://github.com/eugene-panin/ansible-collection-hashistack)
  for the host;
- the OpenTofu module [`eugene-panin/hashistack/nomad`](https://github.com/eugene-panin/terraform-nomad-hashistack)
  for workload identity, Traefik, mail and DNS;
- for OVH, `ovhctl`, a small Go CLI for the OVH API, copied into the repository.

## Use

All you need is Docker. In an empty directory:

```bash
docker run -it --rm -v "$PWD:/work" -v "$HOME/.ssh:/root/.ssh:ro" \
  -v "$HOME/.config:/root/.config" ghcr.io/eugene-panin/hashistack-starter
```

The image asks the questions below, generates the repository in the current
directory, and leaves you in a shell with every tool the next steps need:
Ansible, OpenTofu, Conftest, Go and Boilerplate. Its `README.md` walks through
the rest. Run the same command later to get back into that shell.

Without Docker, install [Boilerplate](https://github.com/gruntwork-io/boilerplate/releases)
0.16 or later and run:

```bash
boilerplate \
  --template-url "github.com/eugene-panin/hashistack-starter//template?ref=v0.2.0" \
  --output-folder ./my-stack
```

It asks:

| Question | Default |
|---|---|
| `ProjectName`: repository name, datacenter name | |
| `Provider`: `ovh` to rebuild an OVH VPS, `ssh` for a server you can log in to | `ovh` |
| `ServerIp`, `BootstrapUser`: the server and the user you log in with today (`ssh`) | `root` |
| `InfraDomain`: domain of the internal names, on Cloudflare | |
| `AcmeEmail`: Let's Encrypt contact | |
| `ServerAddress`, `WireguardNetwork`: the WireGuard /24 | `10.77.0.1`, `10.77.0.0/24` |
| `WireguardClients`: devices that join the tunnel | `[laptop]` |
| `PublicInterface`, `VpsImage`, `OpsUser`, `SshPublicKeyPath` | `ens3` for OVH or `eth0`, `Ubuntu 24.04`, `ops`, `~/.ssh/id_ed25519.pub` |
| `MailEnabled`, `MailHostname`, `MailDomains`, `Mailboxes` | off |
| `ManualDnsDomains`: mail domains whose DNS is not on Cloudflare | `[]` |

Answers can also come from a file, `--var-file answers.yml --non-interactive`,
as in [`test/ssh.yml`](test/ssh.yml).

With `ssh`, `make bootstrap` replaces the OVH rebuild: it logs in as
`BootstrapUser`, creates the ops user with your key, and only then, logged in
as that user, turns off root and password logins, so a key that does not work
stops it before anything is locked. The server is not reinstalled; start from a
fresh Ubuntu 24.04.

## Scope

The internal domain must be on Cloudflare for the wildcard certificate. Mail
domains can be anywhere: those outside Cloudflare get their records printed to
be created by hand.

## Tested

CI builds the image and, inside it, renders the template with
[`test/mail.yml`](test/mail.yml), [`test/minimal.yml`](test/minimal.yml) and
[`test/ssh.yml`](test/ssh.yml). [`test/check.sh`](test/check.sh) checks each
result: shellcheck; `ovhctl` builds and passes its tests, or is absent with
`ssh`; the Conftest policies pass their tests; `make secrets` writes an
encrypted vault and a CA and refuses to run twice; every variable the Ansible
roles read resolves from the generated secrets; the playbooks pass the syntax
check; OpenTofu validates; the Vault init output moves into the vault file.

A second job starts an Ubuntu 24.04 server with systemd that only root can log
in to, and runs the bootstrap from the image against it
([`test/bootstrap.sh`](test/bootstrap.sh)): afterwards the ops user logs in with
the key and uses sudo, root cannot log in, password logins are off, and a
second run changes nothing.

Tags publish the image for amd64 and arm64 to `ghcr.io/eugene-panin/hashistack-starter`.

## License

MIT, see [LICENSE](LICENSE).
