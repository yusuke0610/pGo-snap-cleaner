// Package detect は画像メタデータを分類する。Detector インターフェースが
// 拡張ポイントで、現状は Pokémon GO スナップショット検出器が 1 つだけだが、
// 呼び出し側に手を入れずに新しい検出器を追加できる。
package detect

import (
	"math"

	"github.com/wadayusuke/photo-clean/internal/imagemeta"
)

// Result は 1 枚の画像に対する検出器の判定結果。
type Result struct {
	// Matched は必須ゲートをすべて通過したとき true。
	Matched bool
	// Score はマッチ時の 0..1 の確信度（補助シグナルで加点される）。
	Score float64
	// Reasons はどのゲートを通過/不通過したかを人間向けに説明する。
	Reasons []string
}

// Detector は抽出済みのメタデータを検査し、画像が自分のカテゴリに属するかを返す。
type Detector interface {
	Name() string
	Detect(m imagemeta.Meta) Result
}

// PokemonGO の目標アスペクト比: スナップショットは 690×1227(1×) と
// 1380×2454(2×) で、どちらも厳密に 230:409。よって真の比率は数学的に厳密(差 0)。
// 許容差は 16:9(1.77778, 約 0.00048 差) との隙間より小さく保つこと。さもないと
// EXIF 無しの 16:9 JPEG が誤爆する。0.0003 は実スナップショットに十分な余裕を
// 残しつつ 16:9 を確実に除外する。
const (
	targetRatio = 409.0 / 230.0
	ratioTol    = 0.0003
)

// PokemonGOSnapshot は Pokémon GO の AR スナップショットを 3 つの必須ゲートで
// 検出する: アスペクト比 ≈ 230:409、JPEG コンテナ、カメラ EXIF 不在。
type PokemonGOSnapshot struct{}

func (PokemonGOSnapshot) Name() string { return "pokemon-go-snapshot" }

func (PokemonGOSnapshot) Detect(m imagemeta.Meta) Result {
	var reasons []string

	short := m.ShortSide()
	gateRatio := false
	if short > 0 {
		ratio := float64(m.LongSide()) / float64(short)
		gateRatio = math.Abs(ratio-targetRatio) <= ratioTol
		if gateRatio {
			reasons = append(reasons, "aspect ratio ≈ 230:409")
		} else {
			reasons = append(reasons, "aspect ratio mismatch")
		}
	} else {
		reasons = append(reasons, "no dimensions")
	}

	gateJPEG := m.Format == "jpeg"
	if gateJPEG {
		reasons = append(reasons, "JPEG")
	} else {
		reasons = append(reasons, "not JPEG")
	}

	gateNoEXIF := !m.HasCameraEXIF
	if gateNoEXIF {
		reasons = append(reasons, "no camera EXIF")
	} else {
		reasons = append(reasons, "has camera EXIF")
	}

	if !(gateRatio && gateJPEG && gateNoEXIF) {
		return Result{Matched: false, Reasons: reasons}
	}

	// 全ゲート通過 → high confidence。補助シグナルでスコアを少し上げる。
	score := 0.9
	if m.DPI == 72 {
		score += 0.05
		reasons = append(reasons, "72 DPI")
	}
	if score > 1.0 {
		score = 1.0
	}
	return Result{Matched: true, Score: score, Reasons: reasons}
}
