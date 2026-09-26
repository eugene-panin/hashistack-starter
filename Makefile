.PHONY: test test-mail test-minimal

BOILERPLATE ?= boilerplate

test: test-mail test-minimal

test-mail:
	out=$$(mktemp -d) && $(BOILERPLATE) --template-url template --output-folder $$out --var-file test/mail.yml --non-interactive && test/check.sh $$out

test-minimal:
	out=$$(mktemp -d) && $(BOILERPLATE) --template-url template --output-folder $$out --var-file test/minimal.yml --non-interactive && test/check.sh $$out
