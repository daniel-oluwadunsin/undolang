.DEFAULT_GOAL := help

SHELL := /bin/sh

GO ?= go
BIN ?= dist/undo
FUZZTIME ?= 5s
GO_OFFLINE = GOTOOLCHAIN=go1.27.0 GOPROXY=off
BUILD_ENV = CGO_ENABLED=0 $(GO_OFFLINE)

ROOT_FLAG = $(if $(ROOT),--root "$(ROOT)",)
TRANSACTION_FLAG = $(if $(TRANSACTION),--transaction "$(TRANSACTION)",)
JSON_FLAG = $(if $(JSON),--json,)
YES_FLAG = $(if $(YES),--yes,)
ALLOW_PATH_FLAGS = $(foreach path,$(ALLOW_PATHS),--allow-path "$(path)")

.PHONY: help all build version test vet race fuzz stress crash-test examples \
	check plan run recover history inspect capabilities schema agent-guide \
	modules deps-proof repro release verify

help:
	@echo "UndoLang commands"
	@echo "  make build                                      build $(BIN)"
	@echo "  make version                                    print the binary version"
	@echo "  make test                                       run the offline test suite"
	@echo "  make vet                                        run go vet offline"
	@echo "  make race                                       run the race detector"
	@echo "  make examples                                   parse every examples/*.undo file"
	@echo "  make modules                                    show the module graph (main module only)"
	@echo "  make verify                                     run build, tests, proofs, and release checks"
	@echo "  make check FILE=x.undo ROOT=/path               validate without changing files"
	@echo "  make plan FILE=x.undo ROOT=/path [JSON=1]       inspect the planned changes"
	@echo "  make run FILE=x.undo ROOT=/path YES=1           execute after explicit approval"
	@echo "  make recover ROOT=/path YES=1                   recover an interrupted transaction"
	@echo "  make history ROOT=/path [JSON=1]                list transaction history"
	@echo "  make inspect TXID=id ROOT=/path [JSON=1]        inspect one transaction"
	@echo "  make capabilities                               print the agent capabilities"
	@echo "  make schema                                     print the language schema"
	@echo "  make agent-guide                                print the agent workflow"
	@echo "  make fuzz | make stress | make crash-test       run longer safety checks"
	@echo "  make deps-proof | make repro | make release     write/prove release artifacts"
	@echo ""
	@echo "Optional variables: BIN=..., FUZZTIME=5s, TRANSACTION=..., JSON=1, YES=1, ALLOW_PATHS='dir1 dir2'"

all: build test vet

build:
	mkdir -p "$(dir $(BIN))"
	$(BUILD_ENV) $(GO) build -trimpath -buildvcs=false -o "$(BIN)" ./cmd/undo

version: build
	"$(BIN)" version

test:
	$(GO_OFFLINE) $(GO) test ./...

vet:
	$(GO_OFFLINE) $(GO) vet ./...

race:
	$(GO_OFFLINE) $(GO) test -race ./...

fuzz:
	$(GO_OFFLINE) $(GO) test ./internal/lang/lexer -run '^$$' -fuzz '^FuzzLexNeverPanics$$' -fuzztime $(FUZZTIME)
	$(GO_OFFLINE) $(GO) test ./internal/lang/parser -run '^$$' -fuzz '^FuzzParseNeverPanics$$' -fuzztime $(FUZZTIME)
	$(GO_OFFLINE) $(GO) test ./internal/journal -run '^$$' -fuzz '^FuzzDecodeNeverPanics$$' -fuzztime $(FUZZTIME)

stress:
	UNDOLANG_LARGE_TREE=1 $(GO_OFFLINE) $(GO) test ./internal/plan -run '^TestStressPlan100KEntryTree$$' -count=1
	UNDOLANG_STRESS=1 $(GO_OFFLINE) $(GO) test ./internal/fsop -run '^TestStressCopy256MiB$$' -count=1

crash-test:
	$(GO_OFFLINE) $(GO) test ./internal/cli -run '^TestRealProcessCrashRecoveryMatrix$$' -count=1

examples:
	$(GO_OFFLINE) $(GO) test ./tests -run '^TestEveryExampleParses$$' -count=1

modules:
	$(GO_OFFLINE) $(GO) list -m all

check: build
	@test -n "$(FILE)" || { echo "usage: make check FILE=path/to/file.undo [ROOT=/path] [JSON=1]" >&2; exit 2; }
	"$(BIN)" check "$(FILE)" $(ROOT_FLAG) $(ALLOW_PATH_FLAGS) $(JSON_FLAG)

plan: build
	@test -n "$(FILE)" || { echo "usage: make plan FILE=path/to/file.undo [ROOT=/path] [TRANSACTION=name] [JSON=1]" >&2; exit 2; }
	"$(BIN)" plan "$(FILE)" $(TRANSACTION_FLAG) $(ROOT_FLAG) $(ALLOW_PATH_FLAGS) $(JSON_FLAG)

run: build
	@test -n "$(FILE)" || { echo "usage: make run FILE=path/to/file.undo [ROOT=/path] [TRANSACTION=name] [YES=1] [JSON=1]" >&2; exit 2; }
	"$(BIN)" run "$(FILE)" $(TRANSACTION_FLAG) $(ROOT_FLAG) $(ALLOW_PATH_FLAGS) $(YES_FLAG) $(JSON_FLAG)

recover: build
	@test -n "$(ROOT)" || { echo "usage: make recover ROOT=/path YES=1 [JSON=1]" >&2; exit 2; }
	"$(BIN)" recover $(ROOT_FLAG) $(YES_FLAG) $(JSON_FLAG)

history: build
	@test -n "$(ROOT)" || { echo "usage: make history ROOT=/path [JSON=1]" >&2; exit 2; }
	"$(BIN)" history $(ROOT_FLAG) $(JSON_FLAG)

inspect: build
	@test -n "$(TXID)" || { echo "usage: make inspect TXID=id ROOT=/path [JSON=1]" >&2; exit 2; }
	@test -n "$(ROOT)" || { echo "usage: make inspect TXID=id ROOT=/path [JSON=1]" >&2; exit 2; }
	"$(BIN)" inspect "$(TXID)" $(ROOT_FLAG) $(JSON_FLAG)

capabilities: build
	"$(BIN)" capabilities --json

schema: build
	"$(BIN)" schema --json

agent-guide: build
	"$(BIN)" agent-guide --json

deps-proof:
	./scripts/deps-proof.sh

repro:
	./scripts/repro-build.sh

release:
	./scripts/release.sh

verify: build test vet race examples modules deps-proof repro release
