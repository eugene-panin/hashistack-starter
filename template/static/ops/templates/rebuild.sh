#!/bin/bash
set -euo pipefail

exec > >(tee -a /var/log/ovh-stack-rebuild.log) 2>&1

: "${OPS_USER:?}"
: "${PUBLIC_KEY:?}"
: "${HOST_KEY:?}"
: "${HOST_KEY_PUB:?}"

retry() {
  local attempt=1 max=5
  until "$@"; do
    if [ "$attempt" -ge "$max" ]; then
      echo "giving up after $attempt attempts: $*" >&2
      return 1
    fi
    echo "retry $attempt/$max failed, waiting 5s: $*" >&2
    attempt=$((attempt + 1))
    sleep 5
  done
}

setup_ops_user() {
  useradd -m -s /bin/bash -G sudo "$OPS_USER"
  echo "$OPS_USER ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/90-$OPS_USER"
  chmod 0440 "/etc/sudoers.d/90-$OPS_USER"
  install -d -m 0700 -o "$OPS_USER" -g "$OPS_USER" "/home/$OPS_USER/.ssh"
  printf '%s\n' "$PUBLIC_KEY" > "/home/$OPS_USER/.ssh/authorized_keys"
  chown "$OPS_USER:$OPS_USER" "/home/$OPS_USER/.ssh/authorized_keys"
  chmod 0600 "/home/$OPS_USER/.ssh/authorized_keys"
}

install_host_key() {
  rm -f /etc/ssh/ssh_host_*_key /etc/ssh/ssh_host_*_key.pub
  install -m 0600 /dev/null /etc/ssh/ssh_host_ed25519_key
  printf '%s\n' "$HOST_KEY" > /etc/ssh/ssh_host_ed25519_key
  printf '%s\n' "$HOST_KEY_PUB" > /etc/ssh/ssh_host_ed25519_key.pub
  chmod 0644 /etc/ssh/ssh_host_ed25519_key.pub
}

harden_sshd() {
  cat > /etc/ssh/sshd_config.d/99-hardening.conf <<EOF
HostKey /etc/ssh/ssh_host_ed25519_key
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AllowUsers $OPS_USER
EOF
  retry systemctl restart ssh
}

install_base_packages() {
  retry apt-get update
  retry apt-get install -y ufw unattended-upgrades fail2ban

  cat > /etc/apt/apt.conf.d/20auto-upgrades <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
EOF
}

configure_firewall() {
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow 22/tcp
  ufw allow 80/tcp
  ufw allow 443/tcp
  ufw --force enable
}

main() {
  setup_ops_user
  install_host_key
  harden_sshd
  install_base_packages
  configure_firewall
}

main "$@"
