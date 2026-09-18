GREEN := "\033[0;32m"
NC := "\033[0;0m"

LOG := @sh -c '\
       printf ${GREEN}; \
       echo -e "\n> $$1\n"; \
       printf ${NC}' VALUE

-include .env

MODULE := github.com/auroq/botropolis
BINARIES := botropolis botropolisd botropolis-hook
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X ${MODULE}/pkg/version.Version=${VERSION}
FIXTURE ?= sample

targets ::
	@awk -F'::?[[:space:]]*' '/^[a-zA-Z0-9][^$#\/\t=]*::?([^=]|$$)/ { \
		gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); \
		gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$2); \
		split($$1,A,/ /); \
		target=A[1]; \
		deps=$$2; \
		printf "%s", target; \
		if (deps) printf " → %s", deps; \
		print "" \
	}' $(MAKEFILE_LIST)

build ::
	$(LOG) "Building ${BINARIES} (${VERSION})"
	@for bin in ${BINARIES}; do \
		go build -ldflags "${LDFLAGS}" -o bin/$$bin ./cmd/$$bin || exit 1; \
	done

test :: test-unit test-integration test-acceptance

test-unit ::
	$(LOG) "Running unit tests"
	@go test ./cmd/... ./pkg/...

test-integration :: fixtures
	$(LOG) "Running integration tests"
	@go test ./testing/integration/...

test-acceptance :: fixtures
	$(LOG) "Running acceptance tests"
	@go test ./testing/acceptance/...

lint ::
	$(LOG) "Linting"
	@go vet ./...
	@golangci-lint run ./...

format ::
	$(LOG) "Formatting"
	@gofmt -l -w .

clean ::
	$(LOG) "Cleaning"
	@rm -rf bin dist coverage.out
	@go clean -testcache

analyze ::
	$(LOG) "Analyzing ~/.claude history"
	@python3 tools/analyze-history.py

kits ::
	$(LOG) "Fetching the Kenney kits into tools/kits"
	@tools/fetch-kits

sprites :: kits
	$(LOG) "Cutting the kit atlases into pkg/assets/kits (tools/render-sprites)"
	@blender -b --python tools/render-sprites/render.py -- atlas --out pkg/assets/kits 2>&1 | grep -E "WROTE|BUDGET|Traceback|Error" || true
	@tools/shrink-pngs pkg/assets/kits/kits-z*.png

kit-district :: kits
	$(LOG) "Rendering one district from the Kenney kits (tools/render-sprites)"
	@blender -b --python tools/render-sprites/render.py -- scene --out docs/screenshots/kit-district.png 2>&1 | grep -E "WROTE|Traceback|Error" || true

fixtures ::
	$(LOG) "Generating fixture ${FIXTURE} from ~/.claude"
	@python3 tools/make-fixtures.py --name ${FIXTURE}
