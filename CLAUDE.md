# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## このツールの目的

iPhone の写真ライブラリに紛れ込む **Pokémon GO の AR スナップショット**を判定し、
macOS の Finder タグ `pGo`（赤）を付与するローカル専用 CLI。

**重要な設計原則:**
- **削除機能は作らない。** 判定して印（赤タグ）を付けるだけ。削除はユーザーが Finder で手動。
- **precision 最優先。** 実写真を誤ってタグ付けしないことが最重要。recall より precision。
- **pure Go / CGo 不要 / macOS 専用。** 外部コマンドへのシェルアウトもしない（ファイル内メタだけで判定が成立する）。

## よく使うコマンド

```sh
make build            # ビルド (./photo-clean)
make test             # 全テスト
make test-v           # 詳細表示でテスト
go test ./internal/findertag/ -run TestRemovePGoPreservesOtherTags   # 単一テスト
make vet              # go vet
make fmt              # gofmt -w .
make scan  DIR=~/Pictures REC=1   # 走査のみ（変更なし）
make tag   DIR=~/Pictures REC=1   # pGo(赤)タグ付与
make untag DIR=~/Pictures         # pGo タグ除去
nix build .#default   # Nix で再現ビルド
```

`go` は PATH に無いことがある。Homebrew 版は `/opt/homebrew/bin`、または `nix develop` /
direnv 経由の dev shell に入ると使える。

## アーキテクチャ

3 コマンド（`scan` / `tag` / `untag`）はすべて同じパイプラインを通る。データの流れ:

```
scan.Walk → imagemeta.Extract → detect.Detect → decision.Classify → findertag.{AddPGo,RemovePGo}
（ディレクトリ走査）（ヘッダ/EXIF抽出）（3ゲート判定）（confidence集計） （xattr書き込み）
```

- **internal/scan** — `filepath.WalkDir` でディレクトリ走査。`--recursive` が false なら root 直下のみ。
  画像拡張子でフィルタ（デコード不能な HEIC 等も yield し、判定側で format により落とす）。
- **internal/imagemeta** — `image.DecodeConfig` で**ヘッダのみ**読んで寸法・format を取得（ピクセルはデコードしない）。
  `rwcarlsen/goexif` で EXIF Make/Model の**有無**だけ見る。DPI は任意の補助シグナル。
- **internal/detect** — `Detector` インターフェースが拡張ポイント。現状は `PokemonGOSnapshot` 1 つ。
- **internal/decision** — detector のスコアを confidence ラベル（high/low/none）に変換し、`ShouldTag()` で
  タグ付け可否を決める（full gate match かつ high のときだけタグ）。
- **internal/findertag** — xattr の読み書き + バイナリ plist のエンコード/デコード。
- **cmd/photo-clean** — Cobra の各サブコマンド。各コマンドは上記を組み合わせるだけの薄い層。

### 判定ロジック（最重要・実データで確定済み）

`PokemonGOSnapshot` は次の **3 ゲートすべて通過**で判定する（`internal/detect/detect.go`）:

1. **アスペクト比 ≈ 230:409**（long/short 側で比較。スナップショットは 690×1227 / 1380×2454 で比率は不変・厳密）
2. **format == "jpeg"**（iOS の PNG スクショを除外）
3. **カメラ EXIF（Make/Model）が無い**（似た比率にトリミングした実写真を除外）

注意点:
- **比率の許容差は ±0.0003**（`ratioTol`）。スペック初版は ±0.003 だったが、16:9（1.77778）が
  スナップショット比（1.77826）と 0.00048 しか離れておらず、±0.003 では EXIF 無しの 16:9 JPEG が
  誤爆する。比率は数学的に厳密なので厳しめにして 16:9 を確実に除外している。**緩めないこと。**
- **「カメラ EXIF 不在」を単独の主弁別子にしない。** スクショ・アプリ書き出しも EXIF 不在なので、
  必ず比率ゲートと併用する。
- **機種名で判定しない**（`make == "Apple"` のような決め打ち禁止）。Make/Model が**存在するか否か**で見る
  （一眼レフ写真も弾けるように）。

### Finder タグの仕組み（findertag）

Finder タグは拡張属性 `com.apple.metadata:_kMDItemUserTags` に、`"名前\n色インデックス"` 形式の
文字列配列を**バイナリ plist** で保存する。`pGo` の赤は `"pGo\n6"`（6 = 赤）。

- xattr: `golang.org/x/sys/unix` の `Getxattr`/`Setxattr`/`Removexattr`（2 パスでサイズ確定）
- plist: `howett.net/plist` でバイナリ plist を生成/解析
- **AddPGo は冪等**（既に pGo があれば何もしない）かつ**既存タグを保持**して追記する。
- **RemovePGo は pGo だけを狙い撃ち**で外し、他のタグには絶対触れない。空になれば属性ごと削除。

この「他タグを壊さない」性質は precision と同等に重要。`findertag_test.go` の
`TestRemovePGoPreservesOtherTags` / `TestAddPGoIdempotent` は壊さないこと。

## 開発環境（Nix + direnv）

`flake.nix` が Go ツール一式（go / gopls / gotools / staticcheck / delve）の dev shell を定義。
`.envrc` は nix-direnv の `use flake` を使い、`cd` で自動有効化（初回のみ `direnv allow`）。

`go.mod` / `go.sum` を変更したら `flake.nix` の `vendorHash` を更新する必要がある:
一旦 `lib.fakeHash` に戻して `nix build` → エラーに出る正しいハッシュを貼り替える。

## やらないこと（YAGNI）

検出器は 1 つだけ。汎用判定基盤・YAML 設定・タグの論理演算は作らない。
将来の拡張は `Detector` インターフェース追加で対応する想定（インターフェースだけ拡張可能にしてある）。
