module "stack" {
  source  = "eugene-panin/hashistack/nomad"
  version = "~> 0.6"

  infra_domain             = "{{ .InfraDomain }}"
  address                  = "{{ .ServerAddress }}"
  ca_pem                   = file("${path.root}/../../ansible/files/ca.pem")
  acme_email               = "{{ .AcmeEmail }}"
  dns_provider_env         = { CF_DNS_API_TOKEN = var.traefik_cloudflare_token }
  dns_provider_env_version = var.traefik_cloudflare_token_version
}
{{- if .MailEnabled }}

module "mail" {
  source = "github.com/eugene-panin/terraform-nomad-stalwart?ref=v0.1.0"

  hostname      = "{{ .MailHostname }}"
  domains       = [{{ range $i, $d := .MailDomains }}{{ if $i }}, {{ end }}"{{ $d }}"{{ end }}]
  mailboxes     = [{{ range $i, $m := .Mailboxes }}{{ if $i }}, {{ end }}"{{ $m }}"{{ end }}]
  acme_email    = "{{ .AcmeEmail }}"
  vault_kv_path = module.stack.vault_kv_path
}
{{- end }}

module "dns" {
  source  = "eugene-panin/hashistack/nomad//modules/dns-cloudflare"
  version = "~> 0.6"

  records = [module.stack.dns_records{{ if .MailEnabled }}, module.mail.dns_records{{ end }}]
  domains = ["{{ .InfraDomain }}"{{ if .MailEnabled }}{{ range .MailDomains }}{{ if and (ne . $.InfraDomain) (not (has . $.ManualDnsDomains)) }}, "{{ . }}"{{ end }}{{ end }}{{ end }}]
}
