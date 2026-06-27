// Package findertag は macOS の Finder タグを読み書きする。Finder タグは
// 拡張属性 com.apple.metadata:_kMDItemUserTags に "名前\n色インデックス"
// 形式の文字列配列をバイナリ plist として保存する。本パッケージは "pGo"
// の赤タグ専用で、他のタグには一切触れない。
package findertag

import (
	"bytes"
	"errors"
	"strings"

	"golang.org/x/sys/unix"
	"howett.net/plist"
)

const (
	userTagsAttr = "com.apple.metadata:_kMDItemUserTags"

	// TagName は本ツールが管理する Finder タグ名。
	TagName = "pGo"
	// colorRed は Finder の赤の色インデックス。
	colorRed = 6
)

// pGoTag はディスクに書き込む正規形: 名前 + 改行 + 色。
var pGoTag = TagName + "\n" + string(rune('0'+colorRed))

// ReadTags は path の Finder タグを返す。属性が無い場合はエラーではなく
// 空スライスを返す。
func ReadTags(path string) ([]string, error) {
	data, err := getxattr(path, userTagsAttr)
	if err != nil {
		if errors.Is(err, unix.ENOATTR) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var tags []string
	if _, err := plist.Unmarshal(data, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// HasPGo は path が既に pGo タグ（色は問わない）を持つかを返す。
func HasPGo(path string) (bool, error) {
	tags, err := ReadTags(path)
	if err != nil {
		return false, err
	}
	return indexOfPGo(tags) >= 0, nil
}

// AddPGo は既存タグを保持したまま pGo(赤) タグを追加する。冪等で、既に
// pGo タグがあれば added=false を返して何も書き込まない。
func AddPGo(path string) (bool, error) {
	tags, err := ReadTags(path)
	if err != nil {
		return false, err
	}
	if indexOfPGo(tags) >= 0 {
		return false, nil
	}
	tags = append(tags, pGoTag)
	if err := writeTags(path, tags); err != nil {
		return false, err
	}
	return true, nil
}

// RemovePGo は pGo タグだけを除去し、他のタグには一切触れない。タグが残らない
// 場合は属性ごと削除する。除去対象が無ければ removed=false を返す。
func RemovePGo(path string) (bool, error) {
	tags, err := ReadTags(path)
	if err != nil {
		return false, err
	}
	kept := tags[:0:0]
	removed := false
	for _, t := range tags {
		if tagName(t) == TagName {
			removed = true
			continue
		}
		kept = append(kept, t)
	}
	if !removed {
		return false, nil
	}
	if len(kept) == 0 {
		if err := removexattr(path, userTagsAttr); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := writeTags(path, kept); err != nil {
		return false, err
	}
	return true, nil
}

// indexOfPGo は最初の pGo タグの添字を返す。無ければ -1。
func indexOfPGo(tags []string) int {
	for i, t := range tags {
		if tagName(t) == TagName {
			return i
		}
	}
	return -1
}

// tagName はエンコード済みタグの名前部分を返す（"名前\n色" → "名前"）。
func tagName(tag string) string {
	if i := strings.IndexByte(tag, '\n'); i >= 0 {
		return tag[:i]
	}
	return tag
}

func writeTags(path string, tags []string) error {
	var buf bytes.Buffer
	enc := plist.NewEncoderForFormat(&buf, plist.BinaryFormat)
	if err := enc.Encode(tags); err != nil {
		return err
	}
	return setxattr(path, userTagsAttr, buf.Bytes())
}

// getxattr は拡張属性を読む。バッファサイズを 2 パスで確定する。
func getxattr(path, attr string) ([]byte, error) {
	sz, err := unix.Getxattr(path, attr, nil)
	if err != nil {
		return nil, err
	}
	if sz == 0 {
		return nil, nil
	}
	buf := make([]byte, sz)
	n, err := unix.Getxattr(path, attr, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func setxattr(path, attr string, data []byte) error {
	return unix.Setxattr(path, attr, data, 0)
}

func removexattr(path, attr string) error {
	return unix.Removexattr(path, attr)
}
