GREEN := "\033[0;32m"
NC := "\033[0;0m"

LOG := @sh -c '\
       printf ${GREEN}; \
       echo -e "\n> $$1\n"; \
       printf ${NC}' VALUE

-include .env

MODULE := github.com/auroq/botropolis
BINARIES := botropolis botropolisd botropolis-hook
VERSION ?= $(shell cat VERSION)
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
		go build -trimpath -ldflags "${LDFLAGS}" -o bin/$$bin ./cmd/$$bin || exit 1; \
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

lint :: docs-check version-check
	$(LOG) "Linting"
	@go vet ./...
	@golangci-lint run ./...

docs-check ::
	$(LOG) "Checking the documentation structure"
	@tools/check-docs

version-check ::
	$(LOG) "Checking every file that names a version agrees"
	@tools/check-version

# nfpm builds all three package formats from packaging/nfpm.yaml. It is not in
# Arch's repositories -- `yay -S nfpm-bin` -- and CI installs it on the runner.
nfpm ::
	@command -v nfpm >/dev/null || { echo "nfpm is not installed: yay -S nfpm-bin"; exit 1; }

package-deb :: build nfpm
	$(LOG) "Packaging botropolis ${VERSION} as a deb"
	@mkdir -p dist
	@VERSION=${VERSION} GOARCH=$(shell go env GOARCH) nfpm package --config packaging/nfpm.yaml --packager deb --target dist/

package-rpm :: build nfpm
	$(LOG) "Packaging botropolis ${VERSION} as an rpm"
	@mkdir -p dist
	@VERSION=${VERSION} GOARCH=$(shell go env GOARCH) nfpm package --config packaging/nfpm.yaml --packager rpm --target dist/

package-archlinux :: build nfpm
	$(LOG) "Packaging botropolis ${VERSION} for Arch"
	@mkdir -p dist
	@VERSION=${VERSION} GOARCH=$(shell go env GOARCH) nfpm package --config packaging/nfpm.yaml --packager archlinux --target dist/

package :: package-deb package-rpm package-archlinux

tarball :: build
	$(LOG) "Tarring botropolis ${VERSION} with its licences"
	@mkdir -p dist
	@rm -rf dist/botropolis-${VERSION}
	@mkdir -p dist/botropolis-${VERSION}
	@cp ${BINARIES:%=bin/%} dist/botropolis-${VERSION}/
	@cp LICENSE README.md CHANGELOG.md dist/botropolis-${VERSION}/
	@cp pkg/assets/fonts/inter/LICENSE.txt dist/botropolis-${VERSION}/inter-OFL.txt
	@cp pkg/assets/kits/nature-kit/License.txt dist/botropolis-${VERSION}/kenney-CC0.txt
	@tar -C dist -czf dist/botropolis-${VERSION}-linux-$(shell go env GOARCH).tar.gz botropolis-${VERSION}
	@rm -rf dist/botropolis-${VERSION}

checksums ::
	$(LOG) "Writing SHA256SUMS"
	@cd dist >/dev/null && rm -f SHA256SUMS && sha256sum botropolis-* > SHA256SUMS

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

gif ::
	$(LOG) "Recording docs/botropolis.gif headlessly (24 s, one key every two seconds)"
	@rm -rf dist/frames && mkdir -p dist/frames
	@bin/botropolis city --headless --record dist/frames --seconds 24 --keys n,n,equal,equal,tab,r,r,b,b,x,escape,t,escape,slash,m,a,r,q,enter,f
	@ffmpeg -loglevel error -y -framerate 10 -i dist/frames/frame-%05d.png -vf "fps=8,scale=880:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=160[p];[s1][p]paletteuse=dither=bayer:bayer_scale=4" docs/botropolis.gif
	@ls -la docs/botropolis.gif | awk '{print $$5 " bytes"}'

kits ::
	$(LOG) "Fetching the Kenney kits into tools/kits"
	@tools/fetch-kits

sprites :: kits
	$(LOG) "Cutting the kit atlases into pkg/assets/kits (tools/render-sprites)"
	@blender -b --python tools/render-sprites/render.py -- atlas --out pkg/assets/kits 2>&1 | grep -E "WROTE|BUDGET|Traceback|Error" || true
	@tools/shrink-pngs pkg/assets/kits/kits-z*.png

# Cheap: re-encodes the nine shipped pages and compares, no Blender. This
# is the guard for item 63 — render.py writing the pages and shrink-pngs
# compressing them were two steps with nothing checking that both ran,
# and nine pages shipped unshrunk for fifty revisions because of it.
atlas-shrunk ::
	$(LOG) "Checking every shipped atlas page is at full compression"
	@tools/shrink-pngs --check pkg/assets/kits/kits-z*.png

sprites-check :: sprites atlas-shrunk
	$(LOG) "Checking the re-render against the committed atlases"
	@python3 tools/atlas-diff.py --self-test
	@python3 tools/atlas-diff.py

kit-district :: kits
	$(LOG) "Rendering one district from the Kenney kits (tools/render-sprites)"
	@blender -b --python tools/render-sprites/render.py -- scene --out docs/screenshots/kit-district.png 2>&1 | grep -E "WROTE|Traceback|Error" || true

fixtures ::
	$(LOG) "Generating fixture ${FIXTURE} from ~/.claude"
	@python3 tools/make-fixtures.py --name ${FIXTURE}
