// Package scan はディレクトリツリーを走査し、候補となる画像ファイルを列挙する。
package scan

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// imageExts は検査対象とする拡張子。デコードできないもの（例: .heic）も列挙し、
// 検出器が format で弾く。
var imageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".heic": true,
	".heif": true,
	".tiff": true,
	".tif":  true,
	".bmp":  true,
	".webp": true,
}

// Walk は root 以下の各画像ファイルを訪れ、そのパスで fn を呼ぶ。recursive が
// false のときは root 直下のファイルだけを訪れる。
func Walk(root string, recursive bool, fn func(path string) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if !recursive && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !imageExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		return fn(path)
	})
}
