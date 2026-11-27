# Critter Kingdom developer commands. Run `make` on its own to list them.

GO ?= go
# Build tags, for example TAGS=debug for the debug overlay.
TAGS ?=
# Arguments for the desktop game, for example ARGS=-pad for the touch pad.
ARGS ?=
PORT ?= 8080
# Inkscape 1.x for make art, a name on PATH or a path, for example
# INKSCAPE=/Applications/Inkscape.app/Contents/MacOS/inkscape on a Mac.
INKSCAPE ?= inkscape
# Extra arguments for make art, for example ART_ARGS=-force to export everything.
ART_ARGS ?=

WEB_OUT := build/web
LDFLAGS := -s -w

.PHONY: help run build serve web art test check clean

help: ## List the commands
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN { FS = ":.*## " } { printf "  make %-6s %s\n", $$1, $$2 }'

run: ## Play on the desktop (TAGS=debug for the overlay, ARGS=-pad for the touch pad)
	$(GO) run -tags '$(TAGS)' ./cmd/game $(ARGS)

build: ## Build the desktop game into build/
	$(GO) build -tags '$(TAGS)' -trimpath -o build/critterkingdom ./cmd/game

serve: ## Serve the browser build, rebuilt on every reload
	@echo "Open http://localhost:$(PORT)/web/ here, or this computer's IP address instead of localhost on a tablet."
	$(GO) tool wasmserve -http :$(PORT) -tags '$(TAGS)' ./cmd/game

web: ## Build the browser bundle into build/web
	@mkdir -p $(WEB_OUT)
	GOOS=js GOARCH=wasm $(GO) build -tags '$(TAGS)' -trimpath -ldflags '$(LDFLAGS)' -o $(WEB_OUT)/main.wasm ./cmd/game
	cp -R web/. $(WEB_OUT)/
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(WEB_OUT)/
	sed "s/__WASM_SIZE__/$$(wc -c < $(WEB_OUT)/main.wasm | tr -d ' ')/" web/index.html > $(WEB_OUT)/index.html
	@echo "$(WEB_OUT)/main.wasm is $$(wc -c < $(WEB_OUT)/main.wasm | tr -d ' ') bytes"

art: ## Export changed drawings in assets-src/ to PNG and part JSON (needs Inkscape)
	$(GO) run ./tools/art -inkscape '$(INKSCAPE)' $(ART_ARGS)

test: ## Run the tests
	$(GO) test ./...

check: ## Formatting, go.mod, vet for desktop and browser, and tests
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "gofmt needed:"; echo "$$unformatted"; exit 1; fi
	$(GO) mod tidy -diff
	$(GO) vet ./...
	$(GO) vet -tags debug ./...
	GOOS=js GOARCH=wasm $(GO) vet ./...
	GOOS=js GOARCH=wasm $(GO) vet -tags debug ./...
	$(GO) test ./...

clean: ## Remove build output
	rm -rf build
