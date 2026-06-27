# pGo-snap-cleaner 開発・実行用 Makefile
#
# よく使うコマンド:
#   make build         バイナリをビルド (./pGo-snap-cleaner)
#   make test          テスト実行
#   make scan DIR=...  指定ディレクトリを走査（変更なし）
#   make tag   DIR=... pGo(赤)タグを付与
#   make untag DIR=... pGo タグを除去
#
# DIR / REC は scan/tag/untag に渡す引数。
#   DIR : 対象ディレクトリ（必須）
#   REC : 1 を指定するとサブディレクトリも再帰（-r）

BINARY := pGo-snap-cleaner
PKG    := ./cmd/pGo-snap-cleaner
DIR    ?=
REC    ?=

# REC=1 のとき -r フラグを付ける
RFLAG := $(if $(filter 1,$(REC)),-r,)

.DEFAULT_GOAL := help

## help: このヘルプを表示
.PHONY: help
help:
	@echo "pGo-snap-cleaner — make ターゲット一覧:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

## build: バイナリをビルド (./pGo-snap-cleaner)
.PHONY: build
build:
	go build -o $(BINARY) $(PKG)

## install: $GOBIN (なければ $GOPATH/bin) にインストール
.PHONY: install
install:
	go install $(PKG)

## test: 全テストを実行
.PHONY: test
test:
	go test ./...

## test-v: 詳細表示でテストを実行
.PHONY: test-v
test-v:
	go test -v ./...

## fmt: gofmt で整形
.PHONY: fmt
fmt:
	gofmt -w .

## vet: go vet で静的チェック
.PHONY: vet
vet:
	go vet ./...

## tidy: 依存を整理 (go.mod/go.sum)
.PHONY: tidy
tidy:
	go mod tidy

## scan: ディレクトリを走査して件数を表示（変更なし）。例: make scan DIR=~/Pictures REC=1
.PHONY: scan
scan: build guard-DIR
	./$(BINARY) scan $(RFLAG) $(DIR)

## tag: 判定が pokemon のファイルに pGo(赤)タグを付与。例: make tag DIR=~/Pictures REC=1
.PHONY: tag
tag: build guard-DIR
	./$(BINARY) tag $(RFLAG) $(DIR)

## untag: pGo タグだけを除去。例: make untag DIR=~/Pictures REC=1
.PHONY: untag
untag: build guard-DIR
	./$(BINARY) untag $(RFLAG) $(DIR)

## nix-build: Nix で再現ビルド (./result/bin/pGo-snap-cleaner)
.PHONY: nix-build
nix-build:
	nix build .#default

## clean: ビルド成果物を削除
.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -f result

# DIR が未指定なら分かりやすくエラーにする内部ターゲット
guard-DIR:
	@if [ -z "$(DIR)" ]; then \
		echo "エラー: DIR を指定してください。例: make $(MAKECMDGOALS) DIR=~/Pictures"; \
		exit 1; \
	fi
