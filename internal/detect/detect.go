// Package detect は画像メタデータを分類する。Detector インターフェースが
// 拡張ポイントで、現状は Pokémon GO スナップショット検出器が 1 つだけだが、
// 呼び出し側に手を入れずに新しい検出器を追加できる。
package detect

import (
	"github.com/yusuke0610/pGo-snap-cleaner/internal/imagemeta"
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
// 1380×2454(2×) で、どちらも厳密に 230:409。比率は数学的に厳密なので、
// 浮動小数の許容差ではなく整数で厳密一致を見る（後述の matchesRatio）。
//
// 浮動小数の許容差は危険: 実ライブラリには 1707×960(569:320, 1.778125) という
// 別物の画像があり、230:409(1.778261)との差は 0.000136 しかない。許容差を
// これより小さくしようとすると実用上ほぼゼロになり、結局厳密一致と同じ。
// 整数比較なら誤差ゼロで 1707×960 を確実に除外し、整数倍(k×690 × k×1227)は
// 自動的に拾える。
const (
	ratioW = 230
	ratioH = 409
)

// matchesRatio は long:short が厳密に 409:230 かを整数演算で判定する。
func matchesRatio(long, short int) bool {
	return short > 0 && long*ratioW == short*ratioH
}

// PokemonGOSnapshot は Pokémon GO の AR スナップショットを 3 つの必須ゲートで
// 検出する: アスペクト比 ≈ 230:409、JPEG コンテナ、カメラ EXIF 不在。
type PokemonGOSnapshot struct{}

func (PokemonGOSnapshot) Name() string { return "pokemon-go-snapshot" }

func (PokemonGOSnapshot) Detect(m imagemeta.Meta) Result {
	var reasons []string

	short := m.ShortSide()
	gateRatio := false
	if short > 0 {
		gateRatio = matchesRatio(m.LongSide(), short)
		if gateRatio {
			reasons = append(reasons, "aspect ratio == 230:409")
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
