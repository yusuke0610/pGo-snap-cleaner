// Package imagemeta は画像分類に必要な軽量メタデータ（寸法・コンテナ形式・
// カメラ EXIF(Make/Model) の有無）を抽出する。ピクセルデータはデコードしない。
package imagemeta

import (
	"image"
	// image.DecodeConfig が format と寸法をヘッダのみから読めるよう、
	// デコーダを登録する（ピクセルはデコードしない）。
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

// Meta は検出器が使うヘッダレベルの事実を保持する。
type Meta struct {
	Width  int
	Height int
	// Format は image.DecodeConfig が返す小文字の形式名（例: "jpeg", "png"）。
	// 判定できなかった場合は空。
	Format string
	// HasCameraEXIF は EXIF の Make または Model が非空で存在するとき true。
	// 実写真には必ず入るが、スクショやアプリ書き出しには無い。
	HasCameraEXIF bool
	// DPI は EXIF の水平解像度。取得できなければ 0。任意の補助シグナルのみ。
	DPI int
}

// LongSide は大きい方のピクセル寸法を返す。
func (m Meta) LongSide() int {
	if m.Height >= m.Width {
		return m.Height
	}
	return m.Width
}

// ShortSide は小さい方のピクセル寸法を返す。
func (m Meta) ShortSide() int {
	if m.Height >= m.Width {
		return m.Width
	}
	return m.Height
}

// Extract は Meta を埋めるのに必要なヘッダバイトだけを読む。
func Extract(path string) (Meta, error) {
	f, err := os.Open(path)
	if err != nil {
		return Meta{}, err
	}
	defer f.Close()

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		// 不明・非対応コンテナ（例: HEIC）。分かる範囲を返す。
		// 呼び出し側は jpeg 以外/空の format を非マッチとして扱う。
		return Meta{Format: format}, nil
	}
	m := Meta{Width: cfg.Width, Height: cfg.Height, Format: format}

	// 先頭に巻き戻してカメラ EXIF を探す。カメラ以外の画像はたいてい EXIF を
	// 持たないので、ここでのデコードエラーは単に「EXIF 無し」を意味し致命的ではない。
	if _, err := f.Seek(0, 0); err == nil {
		if x, err := exif.Decode(f); err == nil {
			m.HasCameraEXIF = tagPresent(x, exif.Make) || tagPresent(x, exif.Model)
			if t, err := x.Get(exif.XResolution); err == nil {
				if num, den, err := t.Rat2(0); err == nil && den != 0 {
					m.DPI = int(num / den)
				}
			}
		}
	}
	return m, nil
}

// tagPresent は指定した EXIF タグが存在し非空かどうかを返す。
func tagPresent(x *exif.Exif, field exif.FieldName) bool {
	t, err := x.Get(field)
	if err != nil || t == nil {
		return false
	}
	s, err := t.StringVal()
	if err != nil {
		// タグは存在するが文字列ではない。存在自体を有意とみなす。
		return true
	}
	return s != ""
}
