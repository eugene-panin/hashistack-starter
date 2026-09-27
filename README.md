# hashistack-starter

Turn a freshly rented server into your own small, private cloud, without
learning how to run one.

When you are done:

- your server runs the apps you give it (websites, services) and restarts
  them when they fail;
- it gets HTTPS certificates by itself;
- if you want, it is the mail server for your domains, with a mailbox per
  domain;
- the admin pages open only from your own laptop, through a private
  connection (a VPN); from the internet the server shows nothing but what you
  publish;
- the whole setup is a folder of files on your laptop. Keep it in git, and
  the server can be rebuilt from it at any time.

## What you need

- **A server.**
  - A VPS from any provider, with Ubuntu 24.04, freshly installed.
  - From the provider you need its IP address and a way to log in: the root
    password or your SSH key.
  - 2 CPUs and 4 GB of memory are enough.
- **A domain on Cloudflare.**
  - Cloudflare answers for your domain on the internet; the free plan is
    enough.
  - If your domain is somewhere else: add it in
    [Cloudflare](https://dash.cloudflare.com) ("Add a domain"), then, at the
    company you bought it from, replace its name servers with the two
    Cloudflare shows you.
- **A laptop** with:
  - [Docker Desktop](https://www.docker.com/products/docker-desktop/): all
    the tools run inside it, nothing else to install;
  - the [WireGuard](https://www.wireguard.com/install/) app: the private
    connection to your server.
- **An SSH key.** This is how your laptop proves who it is to the server.
  Check with `ls ~/.ssh/id_ed25519.pub`; if the file is not there, create it
  with `ssh-keygen -t ed25519` and press Enter at every question.

## Start

In a terminal on your laptop, make an empty folder for your setup and start
the starter in it:

```bash
mkdir my-cloud && cd my-cloud
docker run -it --rm -v "$PWD:/work" -v "$HOME/.ssh:/root/.ssh:ro" \
  -v "$HOME/.config:/root/.config" ghcr.io/eugene-panin/hashistack-starter
```

It asks a few questions. The ones that matter:

| Question | Answer |
|---|---|
| `ProjectName` | a short name in lowercase, such as `my-cloud` |
| `Provider` | `ssh` |
| `ServerIp` | the IP address of your server |
| `BootstrapUser` | the user you log in with today, usually `root` |
| `InfraDomain` | a domain on Cloudflare for the admin pages. It must be a whole domain, such as `example-admin.com`, not a name inside one, such as `admin.example.com`. A cheap second domain is best: every name under it that you have not set up points to your private network |
| `AcmeEmail` | your email; Let's Encrypt writes there if a certificate is about to expire |
| `MailEnabled` | `true` if the server should handle your mail |

For every other question, press Enter to keep the default.

The starter then writes your setup into the folder. It leaves you in a
terminal that has every tool the next steps need.

**Open `README.md` in your folder and follow it.** It has your answers filled
in and walks you through the rest, about half an hour:

1. create two Cloudflare keys;
2. generate your passwords;
3. lock the server down;
4. install the platform;
5. connect the VPN;
6. start the services.

To get back into that terminal later, run the same `docker run` command in
the folder again.

## Already have an OVH VPS

Answer `ovh` to `Provider` instead of `ssh`. The starter then reinstalls the
VPS through the OVH API itself, so you do not need its IP address or
password, only an OVH API key. The README in your folder explains how to get
one.

## For engineers

The template is a [Boilerplate](https://github.com/gruntwork-io/boilerplate)
template. It generates a repository for one Ubuntu server running Consul,
Vault and Nomad over WireGuard, with Traefik in front and optionally a
Stalwart mail server. The generated repository holds the answers and the
secrets; the logic lives in published pieces it pins:

- Ansible collections [`eugene_panin.base`](https://github.com/eugene-panin/ansible-collection-base)
  and [`eugene_panin.hashistack`](https://github.com/eugene-panin/ansible-collection-hashistack)
  for the host;
- the OpenTofu module [`eugene-panin/hashistack/nomad`](https://github.com/eugene-panin/terraform-nomad-hashistack)
  for the platform: workload identity, Traefik and DNS;
- with mail, the app module [`eugene-panin/stalwart/nomad`](https://github.com/eugene-panin/terraform-nomad-stalwart),
  the Stalwart mail server as a Nomad job;
- for OVH, `ovhctl`, a small Go CLI for the OVH API, copied into the
  repository.

The image holds Boilerplate, Ansible, OpenTofu, Conftest and Go. Without it,
install Boilerplate 0.16 or later and run:

```bash
boilerplate \
  --template-url "github.com/eugene-panin/hashistack-starter//template?ref=v0.3.0" \
  --output-folder ./my-stack
```

Answers can also come from a file, with `--var-file answers.yml
--non-interactive`, as in [`test/ssh.yml`](test/ssh.yml). The questions and
their defaults are in [`template/boilerplate.yml`](template/boilerplate.yml).

The domain for the internal names must be on Cloudflare, for the wildcard
certificate. Mail domains can be anywhere: those outside Cloudflare get their
records printed, to be created by hand.

### Tested

CI builds the image and renders the template inside it with
[`test/mail.yml`](test/mail.yml), [`test/minimal.yml`](test/minimal.yml) and
[`test/ssh.yml`](test/ssh.yml). [`test/check.sh`](test/check.sh) checks each
result:

- shellcheck;
- `ovhctl` builds and passes its tests, or is absent with `ssh`;
- the Conftest policies pass their tests;
- `make secrets` writes an encrypted vault and a CA, and refuses to run twice;
- every variable the Ansible roles read resolves from the generated secrets;
- the playbooks pass the syntax check;
- OpenTofu validates;
- the Vault init output moves into the vault file.

A second job starts an Ubuntu 24.04 server with systemd that only root can log
in to. It runs the bootstrap from the image against it
([`test/bootstrap.sh`](test/bootstrap.sh)) and checks that afterwards:

- the ops user logs in with the key and uses sudo;
- root cannot log in;
- password logins are off;
- a second run changes nothing.

Not yet tested end to end through the starter: provisioning a real server and
applying the OpenTofu resources.

Tags publish the image for amd64 and arm64 to
`ghcr.io/eugene-panin/hashistack-starter`.

## License

MIT, see [LICENSE](LICENSE).
