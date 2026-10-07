PREFIX ?= /usr/local
SCRIPTS := $(wildcard bin/*)

.PHONY: help test lint status report rotate install

help:
	@echo "targets: test lint status report rotate install"

test:
	@bash test/run.sh

lint:
	@shellcheck -x $(SCRIPTS) lib/common.sh test/run.sh

status:
	@bin/beacon-status || true

report:
	@bin/beacon-report

rotate:
	@bin/beacon-rotate --dry-run

install:
	install -d $(DESTDIR)$(PREFIX)/lib/beacon/bin $(DESTDIR)$(PREFIX)/lib/beacon/lib
	install -m 0755 $(SCRIPTS) $(DESTDIR)$(PREFIX)/lib/beacon/bin/
	install -m 0644 lib/common.sh $(DESTDIR)$(PREFIX)/lib/beacon/lib/
