# hashistack-starter

A [Boilerplate](https://github.com/gruntwork-io/boilerplate) template that
generates a repository for one OVH VPS running Consul, Vault and Nomad over
WireGuard, with Traefik in front and, if you want it, a mail server. The
generated repository holds your answers and your secrets; the logic lives in
published pieces it pins:

- Ansible collections [`eugene_panin.base`](https://github.com/eugene-panin/ansible-collection-base)
  and [`eugene_panin.hashistack`](https://github.com/eugene-panin/ansible-collection-hashistack)
  for the host;
- the OpenTofu module [`eugene-panin/hashistack/nomad`](https://github.com/eugene-panin/terraform-nomad-hashistack)
  for workload identity, Traefik, mail and DNS;
- `ovhctl`, a small Go CLI for the OVH API, copied into the repository.

## Use

Install [Boilerplate](https://github.com/gruntwork-io/boilerplate/releases)
0.16 or later, then:

```bash
boilerplate \
  --template-url "github.com/eugene-panin/hashistack-starter//template?ref=v0.1.0" \
  --output-folder ./my-stack
```

It asks:

| Question | Default |
|---|---|
| `ProjectName`: repository name, datacenter name | |
| `InfraDomain`: domain of the internal names, on Cloudflare | |
| `AcmeEmail`: Let's Encrypt contact | |
| `ServerAddress`, `WireguardNetwork`: the WireGuard /24 | `10.77.0.1`, `10.77.0.0/24` |
| `WireguardClients`: devices that join the tunnel | `[laptop]` |
| `PublicInterface`, `VpsImage`, `OpsUser`, `SshPublicKeyPath` | `ens3`, `Ubuntu 24.04`, `ops`, `~/.ssh/id_ed25519.pub` |
| `MailEnabled`, `MailHostname`, `MailDomains`, `Mailboxes` | off |
| `ManualDnsDomains`: mail domains whose DNS is not on Cloudflare | `[]` |

Answers can also come from a file, `--var-file answers.yml --non-interactive`,
as in [`test/mail.yml`](test/mail.yml). The generated `README.md` walks through
the rest: credentials, `make secrets`, rebuild, provision, OpenTofu.

## Scope

OVH VPS only for now; the internal domain must be on Cloudflare for the
wildcard certificate. Mail domains can be anywhere: those outside Cloudflare
get their records printed to be created by hand.

## Tested

CI renders the template with [`test/mail.yml`](test/mail.yml) and
[`test/minimal.yml`](test/minimal.yml), and [`test/check.sh`](test/check.sh)
checks each result: shellcheck, `ovhctl` builds and passes its tests, the
Conftest policies pass their tests, `make secrets` writes an encrypted vault and
a CA and refuses to run twice, every variable the Ansible roles read resolves
from the generated secrets, the playbook passes the syntax check, OpenTofu
validates, and the Vault init output moves into the vault file.

## License

MIT, see [LICENSE](LICENSE).
