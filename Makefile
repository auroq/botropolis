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

# Ebitengine reaches GL and X11 through dlopen rather than linking them, so
# nothing here needs cgo -- verified by building with it off and rendering a
# frame. That is what makes GOARCH=arm64 a plain `go build` with no cross
# toolchain, and it is why the release matrix is cheap.
export CGO_ENABLED = 0
GOARCH ?= $(shell go env GOARCH)

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

# The demo stager is a development tool for filming the city, so it is built
# beside the binaries rather than among them: nothing packages it.
demo-tool ::
	$(LOG) "Building botropolis-demo"
	@go build -trimpath -ldflags "${LDFLAGS}" -o bin/botropolis-demo ./cmd/botropolis-demo

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

lint :: docs-check version-check workflows-check
	$(LOG) "Linting"
	@go vet ./...
	@golangci-lint run ./...

docs-check ::
	$(LOG) "Checking the documentation structure"
	@tools/check-docs

version-check ::
	$(LOG) "Checking every file that names a version agrees"
	@tools/check-version

workflows-check ::
	$(LOG) "Checking the workflows parse"
	@tools/check-workflows

packages-check ::
	$(LOG) "Checking the three packages carry the same files"
	@tools/check-packages

# Deliberately not part of `lint`: it reaches the network to ask what the remote
# advertises, and it is a gate for making the repository public rather than a
# per-commit check.
publishable ::
	$(LOG) "Checking nothing unpublishable is in the repository"
	@tools/check-publishable

# nfpm builds all three package formats from packaging/nfpm.yaml. It is not in
# Arch's repositories -- `yay -S nfpm-bin` -- and CI installs it on the runner.
nfpm ::
	@command -v nfpm >/dev/null || { echo "nfpm is not installed: yay -S nfpm-bin"; exit 1; }

package-deb :: build nfpm
	$(LOG) "Packaging botropolis ${VERSION} as a deb"
	@mkdir -p dist
	@VERSION=${VERSION} GOARCH=${GOARCH} nfpm package --config packaging/nfpm.yaml --packager deb --target dist/

package-rpm :: build nfpm
	$(LOG) "Packaging botropolis ${VERSION} as an rpm"
	@mkdir -p dist
	@VERSION=${VERSION} GOARCH=${GOARCH} nfpm package --config packaging/nfpm.yaml --packager rpm --target dist/

package-archlinux :: build nfpm
	$(LOG) "Packaging botropolis ${VERSION} for Arch"
	@mkdir -p dist
	@VERSION=${VERSION} GOARCH=${GOARCH} nfpm package --config packaging/nfpm.yaml --packager archlinux --target dist/

package :: package-deb package-rpm package-archlinux

# The units, desktop entry, icon and shell shim travel with the binaries, so a
# package built from this tarball installs the same set as one built from source.
# Without them botropolis-bin could only ship three executables.
tarball :: build
	$(LOG) "Tarring botropolis ${VERSION} with its licences and packaging"
	@mkdir -p dist
	@rm -rf dist/botropolis-${VERSION}
	@mkdir -p dist/botropolis-${VERSION}
	@cp ${BINARIES:%=bin/%} dist/botropolis-${VERSION}/
	@cp LICENSE README.md CHANGELOG.md dist/botropolis-${VERSION}/
	@cp pkg/assets/fonts/inter/LICENSE.txt dist/botropolis-${VERSION}/inter-OFL.txt
	@cp pkg/assets/kits/nature-kit/License.txt dist/botropolis-${VERSION}/kenney-CC0.txt
	@cp pkg/assets/kits/README.md dist/botropolis-${VERSION}/kenney-kits.md
	@cp pkg/assets/kenney/README.md dist/botropolis-${VERSION}/kenney-packs.md
	@cp -r packaging dist/botropolis-${VERSION}/packaging
	@rm -f dist/botropolis-${VERSION}/packaging/nfpm.yaml
	@tar -C dist -czf dist/botropolis-${VERSION}-linux-${GOARCH}.tar.gz botropolis-${VERSION}
	@rm -rf dist/botropolis-${VERSION}

# botropolis-* misses the deb, which is botropolis_0.1.0_amd64.deb by Debian
# convention. Everything in dist/ is checksummed, and the count is asserted so a
# naming convention nobody expected cannot quietly drop a file again.
checksums ::
	$(LOG) "Writing SHA256SUMS"
	@cd dist >/dev/null && rm -f SHA256SUMS && sha256sum $$(ls | grep -v '^SHA256SUMS$$') > SHA256SUMS
	@cd dist >/dev/null && n=$$(ls | grep -cv '^SHA256SUMS$$') && m=$$(wc -l < SHA256SUMS) && \
		[ "$$n" = "$$m" ] || { echo "SHA256SUMS covers $$m of $$n files in dist/"; exit 1; }

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

# Recorded from the scrubbed fixture, not from the machine it is run on. The
# committed GIF used to be of the author's own city, so it carried project names
# on the district plates, MCP vendor names on the towers and a real daily spend
# across the strip -- published, in a repository whose history had just been
# rewritten to remove exactly that. Pointing this at ~/.claude is opt-in now.
gif :: fixtures
	$(LOG) "Recording docs/botropolis.gif headlessly from the ${FIXTURE} fixture (24 s)"
	@rm -rf dist/frames && mkdir -p dist/frames
	@bin/botropolis city --headless --record dist/frames --seconds 24 \
		--home testing/helpers/fixtures/${FIXTURE}/home --socket /nonexistent/botropolis.sock \
		--keys n,n,equal,equal,tab,r,r,b,b,x,escape,t,escape,slash,m,a,r,q,enter,f
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
