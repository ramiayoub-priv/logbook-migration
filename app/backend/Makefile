# Logbook backend.
#
# The coverage bar is CLAUDE.md rule 0.6 and is enforced here, not by good
# intentions: >=80% overall, 100% on the calculation core.

GO      ?= go
BIN     := dist/logbook-server
PKG     := ./...

# Packages where a bug means a wrong legal record -- or an exposed one. These
# must be at 100%.
#
# internal/auth and internal/ratelimit are here for the second reason.
# CLAUDE.md rule 0.6 names the calculation core, but app/docs/security.md
# requires every control to have a test that fails if the control is removed,
# and in a credential primitive an untested branch is an authentication bypass.
# Same bar, different failure.
CORE    := internal/hhmm internal/timeutil internal/csvbook internal/stats \
           internal/auth internal/ratelimit

MIN_TOTAL := 80

.PHONY: all
all: test build

.PHONY: test
test:
	$(GO) test -race $(PKG)

.PHONY: cover
cover:
	@$(GO) test -coverprofile=coverage.out $(PKG) >/dev/null
	@$(GO) tool cover -func=coverage.out | tail -1
	@total=$$($(GO) tool cover -func=coverage.out | tail -1 | grep -oE '[0-9]+\.[0-9]+' ); \
	 pass=$$(echo "$$total >= $(MIN_TOTAL)" | bc -l); \
	 if [ "$$pass" != "1" ]; then \
	   echo "FAIL: total coverage $$total%% is below the $(MIN_TOTAL)%% bar"; exit 1; \
	 fi
	@echo "OK: total coverage meets the $(MIN_TOTAL)% bar"

# Each core package must be at exactly 100%. Checked per package so that a
# well-covered package cannot mask a thin one behind a healthy average.
.PHONY: cover-core
cover-core:
	@fail=0; \
	for p in $(CORE); do \
	  if ! ls $$p/*.go >/dev/null 2>&1; then echo "SKIP $$p (not written yet)"; continue; fi; \
	  pct=$$($(GO) test -cover ./$$p 2>/dev/null | grep -oE 'coverage: [0-9.]+' | grep -oE '[0-9.]+'); \
	  if [ "$$pct" = "100.0" ]; then echo "OK   $$p 100.0%"; \
	  else echo "FAIL $$p $${pct:-no}% (calculation core must be 100%)"; fail=1; fi; \
	done; \
	exit $$fail

.PHONY: check
check: vet test cover cover-core

.PHONY: vet
vet:
	$(GO) vet $(PKG)

# Cross-compiled for the deploy target. CGO_ENABLED=0 is what makes the result
# a static binary needing no toolchain, libc or Go install on the server --
# modernc.org/sqlite is pure Go precisely so this works.
#
# Every command under cmd/ is built. ./cmd/server arrives with Task 4;
# ./cmd/logbookctl is the operator tool that imports the CSVs.
.PHONY: build
build:
	@mkdir -p dist
	@for c in cmd/*/; do \
	  name=$$(basename $$c); \
	  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	    $(GO) build -trimpath -ldflags="-s -w" -o dist/$$name ./$$c || exit 1; \
	done
	@ls -lh dist/

.PHONY: clean
clean:
	rm -rf dist coverage.out
