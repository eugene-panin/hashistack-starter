output "ui_urls" {
  description = "Addresses of the Consul, Nomad and Vault UIs."
  value       = module.stack.ui_urls
}

output "mailboxes" {
  description = "Mailboxes and the aliases each one receives."
  value       = module.stack.mailboxes
}

output "mail_passwords" {
  description = "Password of each mailbox."
  value       = module.stack.mail_passwords
  sensitive   = true
}

output "mail_manual_records" {
  description = "DNS records to create by hand, for the domains outside Cloudflare."
  value       = { for d, records in module.stack.dns_records : d => records if !contains(keys(module.dns.zone_ids), d) }
}
