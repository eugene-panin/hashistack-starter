.PHONY: test test-mail test-minimal test-ssh

BOILERPLATE ?= boilerplate

test: test-mail test-minimal test-ssh

test-mail test-minimal test-ssh: test-%:
	out=$$(mktemp -d) && $(BOILERPLATE) --template-url template --output-folder $$out --var-file test/$*.yml --non-interactive && test/check.sh $$out
