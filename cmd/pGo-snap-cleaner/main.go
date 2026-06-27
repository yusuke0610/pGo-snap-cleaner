// Command pGo-snap-cleaner は写真ライブラリから Pokémon GO の AR スナップショットを
// 走査し、Finder タグ "pGo"（赤）を付与する。ファイルは削除しない。
package main

import (
	"fmt"
	"os"
)

// version はリリースビルド時に GoReleaser の ldflags で注入される
// （-X main.version=...）。ソースから直接ビルドした場合は "dev"。
var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
