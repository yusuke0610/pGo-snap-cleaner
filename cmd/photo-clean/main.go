// Command photo-clean は写真ライブラリから Pokémon GO の AR スナップショットを
// 走査し、Finder タグ "pGo"（赤）を付与する。ファイルは削除しない。
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
