FROM geerlingguy/docker-ubuntu2404-ansible:latest

RUN apt-get update \
 && apt-get install -y --no-install-recommends openssh-server sudo \
 && rm -rf /var/lib/apt/lists/* \
 && systemctl enable ssh \
 && install -d -m 0700 /root/.ssh

COPY authorized_keys /root/.ssh/authorized_keys
RUN chmod 0600 /root/.ssh/authorized_keys
