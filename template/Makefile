.PHONY: setup secrets vault-import-init check-infra \
        infra-fmt infra-init infra-validate infra-plan infra-policy infra-apply infra-output policy-test \
{{- if eq .Provider "ovh" }}
        check-ops build vps-list vps-use vps-images vps-rebuild vps-host vps-console vps-reboot vps-start vps-stop \
{{- else }}
        bootstrap \
{{- end }}
        provision-deps provision-check provision vault-edit vault-view

INFRA_DIR   := infra/opentofu
POLICY_DIR  := policy
ANSIBLE_DIR := ansible

VAULT_PASSWORD_FILE ?= $(HOME)/.config/{{ .ProjectName }}/vault-pass
VAULT_FILE ?= inventory/group_vars/stack/vault.yml

export ANSIBLE_VAULT_PASSWORD_FILE := $(VAULT_PASSWORD_FILE)
export STACK_ADDRESS := {{ .ServerAddress }}
{{- if eq .Provider "ovh" }}

OPS_DIR      := ops
SERVICE_NAME ?=
IMAGE_ID     ?=
IMAGE_NAME   ?=
PUBLIC_KEY   ?= {{ .SshPublicKeyPath }}
SHOW_SCRIPT  ?=

SERVICE_NAME_ARGS := $(if $(SERVICE_NAME),--service-name $(SERVICE_NAME),)
STACK_HOST = $$(cd $(CURDIR)/$(OPS_DIR) && ../bin/ovhctl host $(SERVICE_NAME_ARGS) 2>/dev/null)
PROVISION_DEPS := build
SECRET_FILES := secrets/ovh-api.env secrets/cloudflare.env
{{- else }}

ASK_PASS ?=

STACK_HOST := {{ .ServerIp }}
PROVISION_DEPS :=
SECRET_FILES := secrets/cloudflare.env
{{- end }}

setup:
	@for f in $(SECRET_FILES); do \
		if [ -f $$f ]; then echo "$$f already exists, left untouched" >&2; \
		else cp $$f.example $$f && chmod 600 $$f && echo "created $$f, fill it in" >&2; fi; \
	done

secrets:
	bin/generate-secrets

vault-import-init:
	bin/import-vault-init

check-infra:
	@test -d $(INFRA_DIR) || { echo "$(INFRA_DIR) is missing" >&2; exit 1; }

infra-fmt: check-infra
	cd $(INFRA_DIR) && tofu fmt

infra-init: check-infra
	$(INFRA_DIR)/tofu.sh init

infra-validate: check-infra
	$(INFRA_DIR)/tofu.sh validate

infra-plan: check-infra
	$(INFRA_DIR)/tofu.sh plan -out=infra.tfplan
	$(POLICY_DIR)/check.sh infra.tfplan

infra-policy: check-infra
	$(POLICY_DIR)/check.sh infra.tfplan

infra-apply: check-infra
	@test "$(CONFIRM)" = "1" || { echo "infra-apply requires CONFIRM=1" >&2; exit 1; }
	$(POLICY_DIR)/check.sh infra.tfplan
	$(INFRA_DIR)/tofu.sh apply infra.tfplan

infra-output: check-infra
	$(INFRA_DIR)/tofu.sh output $(OUTPUT)

policy-test:
	conftest verify --policy $(POLICY_DIR) --data $(POLICY_DIR)/data
{{- if eq .Provider "ovh" }}

check-ops:
	@test -d $(OPS_DIR) || { echo "$(OPS_DIR) is missing" >&2; exit 1; }

build: check-ops
	cd $(OPS_DIR) && go build -o ../bin/ovhctl .

vps-list: build
	cd $(OPS_DIR) && ../bin/ovhctl list

vps-use: build
	cd $(OPS_DIR) && ../bin/ovhctl use $(SERVICE_NAME_ARGS)

vps-images: build
	cd $(OPS_DIR) && ../bin/ovhctl images $(SERVICE_NAME_ARGS)

REBUILD_ARGS := $(SERVICE_NAME_ARGS) $(if $(IMAGE_ID),--image-id $(IMAGE_ID),) $(if $(IMAGE_NAME),--image-name "$(IMAGE_NAME)",) --public-key $(PUBLIC_KEY) $(if $(filter 1,$(CONFIRM)),--confirm,) $(if $(filter 1,$(SHOW_SCRIPT)),--show-script,)

vps-rebuild: build
	cd $(OPS_DIR) && ../bin/ovhctl rebuild $(REBUILD_ARGS)

vps-host: build
	@cd $(OPS_DIR) && ../bin/ovhctl host $(SERVICE_NAME_ARGS)

vps-console: build
	cd $(OPS_DIR) && ../bin/ovhctl console $(SERVICE_NAME_ARGS)

vps-reboot: build
	@test "$(CONFIRM)" = "1" || { echo "vps-reboot requires CONFIRM=1" >&2; exit 1; }
	cd $(OPS_DIR) && ../bin/ovhctl reboot $(SERVICE_NAME_ARGS)

vps-start: build
	cd $(OPS_DIR) && ../bin/ovhctl start $(SERVICE_NAME_ARGS)

vps-stop: build
	@test "$(CONFIRM)" = "1" || { echo "vps-stop requires CONFIRM=1" >&2; exit 1; }
	cd $(OPS_DIR) && ../bin/ovhctl stop $(SERVICE_NAME_ARGS)
{{- else }}

bootstrap: provision-deps
	cd $(ANSIBLE_DIR) && STACK_HOST="$(STACK_HOST)" ansible-playbook playbooks/bootstrap.yml \
	  -e '{"ansible_ssh_common_args": "-o StrictHostKeyChecking=accept-new"}' \
	  $(if $(filter 1,$(ASK_PASS)),--ask-pass --ask-become-pass,)
{{- end }}

provision-deps:
	cd $(ANSIBLE_DIR) && ansible-galaxy collection install -r requirements.yml -p .collections

provision-check: $(PROVISION_DEPS)
	cd $(ANSIBLE_DIR) && STACK_HOST="$(STACK_HOST)" ansible-playbook playbooks/provision.yml --check --diff

provision: $(PROVISION_DEPS)
	cd $(ANSIBLE_DIR) && STACK_HOST="$(STACK_HOST)" ansible-playbook playbooks/provision.yml

vault-edit:
	cd $(ANSIBLE_DIR) && ansible-vault edit $(VAULT_FILE)

vault-view:
	cd $(ANSIBLE_DIR) && ansible-vault view $(VAULT_FILE)
