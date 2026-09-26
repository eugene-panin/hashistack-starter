FROM geerlingguy/docker-ubuntu2404-ansible:latest

RUN apt-get update \
 && apt-get install -y --no-install-recommends openssh-server sudo \
 && systemctl enable ssh \
 && install -d -m 0700 /root/.ssh

COPY authorized_keys /root/.ssh/authorized_keys
RUN chmod 0600 /root/.ssh/authorized_keys
