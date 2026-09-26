FROM python:3.13-slim-bookworm

ARG TARGETARCH
ARG ANSIBLE_CORE_VERSION=2.21.4
ARG OPENTOFU_VERSION=1.12.5
ARG CONFTEST_VERSION=0.70.1
ARG BOILERPLATE_VERSION=0.16.0
ARG GO_VERSION=1.26.6

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl git jq make openssh-client openssl shellcheck sshpass unzip \
 && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir "ansible-core==${ANSIBLE_CORE_VERSION}" pyyaml

RUN set -eu; \
    case "$TARGETARCH" in \
      amd64) \
        tofu_sha=dade9650e6b74fc7a8b986bd8717497d32f9e09cf82e479afef4977fa3085536; \
        conftest_arch=x86_64; conftest_sha=613d124b8f6c1f3cee890491f7ab19114cca5a2102ca47cb2e6c35b4c23f9c8a; \
        boilerplate_sha=e34f6c1ec7af9267703bf2d910f49f4f603a7b80876889eed38db0a751100a1a; \
        go_sha=708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89 ;; \
      arm64) \
        tofu_sha=528f4eea63452bbddb30fa4f1780b57fac8d7676f9dda0f772e847bb62c1260a; \
        conftest_arch=arm64; conftest_sha=8eb914755cb1b3c610d557019d5f373d5528b366ba706961d1a23f2ec55eab57; \
        boilerplate_sha=cce80725952643db8af84ba3d3e3feba21cde9dbd6f3f34c16e20c7f554712f5; \
        go_sha=d0507e9e9d7fe012aae570108cbd76c15de879e17130ab8cb90d4d7445cb1f2e ;; \
      *) echo "unsupported architecture $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    cd /tmp; \
    curl -fsSLo tofu.zip "https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.zip"; \
    echo "$tofu_sha  tofu.zip" | sha256sum -c -; \
    unzip -q tofu.zip tofu -d /usr/local/bin; \
    curl -fsSLo conftest.tar.gz "https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_Linux_${conftest_arch}.tar.gz"; \
    echo "$conftest_sha  conftest.tar.gz" | sha256sum -c -; \
    tar -xzf conftest.tar.gz -C /usr/local/bin conftest; \
    curl -fsSLo /usr/local/bin/boilerplate "https://github.com/gruntwork-io/boilerplate/releases/download/v${BOILERPLATE_VERSION}/boilerplate_linux_${TARGETARCH}"; \
    echo "$boilerplate_sha  /usr/local/bin/boilerplate" | sha256sum -c -; \
    chmod 0755 /usr/local/bin/boilerplate; \
    curl -fsSLo go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz"; \
    echo "$go_sha  go.tar.gz" | sha256sum -c -; \
    tar -xzf go.tar.gz -C /usr/local; \
    rm -f tofu.zip conftest.tar.gz go.tar.gz

ENV PATH=/usr/local/go/bin:$PATH GOTOOLCHAIN=local

COPY template /opt/hashistack-starter/template
COPY image/entrypoint /usr/local/bin/entrypoint

WORKDIR /work
ENTRYPOINT ["/usr/local/bin/entrypoint"]
