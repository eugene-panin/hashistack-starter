module "stack" {
  source  = "eugene-panin/hashistack/nomad"
  version = "~> 0.4"

  infra_domain             = "{{ .InfraDomain }}"
  address                  = "{{ .ServerAddress }}"
  ca_pem                   = file("${path.root}/../../ansible/files/ca.pem")
  acme_email               = "{{ .AcmeEmail }}"
  dns_provider_env         = { CF_DNS_API_TOKEN = var.traefik_cloudflare_token }
  dns_provider_env_version = var.traefik_cloudflare_token_version
{{- if .MailEnabled }}

  mail = {
    hostname  = "{{ .MailHostname }}"
    domains   = [{{ range $i, $d := .MailDomains }}{{ if $i }}, {{ end }}"{{ $d }}"{{ end }}]
    mailboxes = [{{ range $i, $m := .Mailboxes }}{{ if $i }}, {{ end }}"{{ $m }}"{{ end }}]
  }
{{- end }}
}

module "dns" {
  source  = "eugene-panin/hashistack/nomad//modules/dns-cloudflare"
  version = "~> 0.4"

  records = module.stack.dns_records
  domains = ["{{ .InfraDomain }}"{{ if .MailEnabled }}{{ range .MailDomains }}{{ if and (ne . $.InfraDomain) (not (has . $.ManualDnsDomains)) }}, "{{ . }}"{{ end }}{{ end }}{{ end }}]
}
