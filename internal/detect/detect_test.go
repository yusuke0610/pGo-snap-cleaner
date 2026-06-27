package detect

import (
	"testing"

	"github.com/yusuke0610/pGo-snap-cleaner/internal/imagemeta"
)

func TestPokemonGOSnapshot(t *testing.T) {
	d := PokemonGOSnapshot{}

	cases := []struct {
		name string
		m    imagemeta.Meta
		want bool
	}{
		{
			name: "1x スナップショット 690x1227",
			m:    imagemeta.Meta{Width: 690, Height: 1227, Format: "jpeg"},
			want: true,
		},
		{
			name: "2x スナップショット 1380x2454",
			m:    imagemeta.Meta{Width: 1380, Height: 2454, Format: "jpeg"},
			want: true,
		},
		{
			name: "実写真: 比率は正しいがカメラ EXIF を持つ",
			m:    imagemeta.Meta{Width: 690, Height: 1227, Format: "jpeg", HasCameraEXIF: true},
			want: false,
		},
		{
			name: "iOS スクショ: PNG・EXIF 無しだが形式違い",
			m:    imagemeta.Meta{Width: 1170, Height: 2532, Format: "png"},
			want: false,
		},
		{
			name: "EXIF 無し JPEG だが比率違い(16:9 の動画フレーム)",
			m:    imagemeta.Meta{Width: 1080, Height: 1920, Format: "jpeg"},
			want: false,
		},
		{
			name: "誤検出していた 1707x960 (569:320, 比率は近いが厳密には非一致)",
			m:    imagemeta.Meta{Width: 1707, Height: 960, Format: "jpeg"},
			want: false,
		},
		{
			name: "誤検出していた 960x1707 (縦向きでも除外)",
			m:    imagemeta.Meta{Width: 960, Height: 1707, Format: "jpeg"},
			want: false,
		},
		{
			name: "整数倍の 3x スナップ 2070x3681 も拾う",
			m:    imagemeta.Meta{Width: 2070, Height: 3681, Format: "jpeg"},
			want: true,
		},
		{
			name: "寸法がゼロ",
			m:    imagemeta.Meta{Format: "jpeg"},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := d.Detect(tc.m)
			if got.Matched != tc.want {
				t.Fatalf("Matched=%v want %v (reasons: %v)", got.Matched, tc.want, got.Reasons)
			}
			if got.Matched && got.Score < 0.9 {
				t.Fatalf("matched snapshot should score >=0.9, got %v", got.Score)
			}
		})
	}
}

func TestDPIBonus(t *testing.T) {
	d := PokemonGOSnapshot{}
	base := d.Detect(imagemeta.Meta{Width: 690, Height: 1227, Format: "jpeg"})
	bonus := d.Detect(imagemeta.Meta{Width: 690, Height: 1227, Format: "jpeg", DPI: 72})
	if !(bonus.Score > base.Score) {
		t.Fatalf("72 DPI should raise score: base=%v bonus=%v", base.Score, bonus.Score)
	}
}
