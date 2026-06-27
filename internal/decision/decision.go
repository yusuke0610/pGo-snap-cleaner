// Package decision はメタデータ抽出と検出を結びつけ、生のスコアを confidence
// ラベルとタグ付け可否の判断に変換する。
package decision

import (
	"github.com/wadayusuke/photo-clean/internal/detect"
	"github.com/wadayusuke/photo-clean/internal/imagemeta"
)

// Confidence は判定結果の粗いラベル。
type Confidence string

const (
	None Confidence = "none"
	Low  Confidence = "low"
	High Confidence = "high"
)

// Outcome は 1 ファイルに対する完全な判定結果。
type Outcome struct {
	Path       string
	Meta       imagemeta.Meta
	Result     detect.Result
	Confidence Confidence
}

// ShouldTag はこの結果がタグ付けに値するかを返す。全ゲート通過のときだけ
// タグ付けする（recall より precision を優先し、実写真は絶対にタグ付けしない）。
func (o Outcome) ShouldTag() bool {
	return o.Result.Matched && o.Confidence == High
}

// Classifier は 1 ファイルに対して検出器を実行する。
type Classifier struct {
	Detector detect.Detector
}

// New は Pokémon GO スナップショット検出器を使う Classifier を返す。
func New() Classifier {
	return Classifier{Detector: detect.PokemonGOSnapshot{}}
}

// Classify はメタデータを抽出し、検出器を適用する。
func (c Classifier) Classify(path string) (Outcome, error) {
	m, err := imagemeta.Extract(path)
	if err != nil {
		return Outcome{Path: path}, err
	}
	res := c.Detector.Detect(m)
	return Outcome{
		Path:       path,
		Meta:       m,
		Result:     res,
		Confidence: confidenceFor(res),
	}, nil
}

func confidenceFor(r detect.Result) Confidence {
	if !r.Matched {
		return None
	}
	if r.Score >= 0.9 {
		return High
	}
	return Low
}
