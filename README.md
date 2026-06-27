# photo-clean

iPhone の写真ライブラリに紛れ込む **Pokémon GO の AR スナップショット**を自動判定し、
macOS の Finder タグ **`pGo`（赤）** を付与するローカル専用 CLI です。

**このツールはファイルを削除しません。** 印（赤タグ）を付けるだけです。
実際の削除は、Finder 上で赤タグを目視確認しながら手動で行ってください。

macOS 専用 / pure Go（CGo 不要）/ 単一バイナリ。

## 判定ロジック

次の **3 つのゲートをすべて通過**したものだけを Pokémon GO スナップショットと判定します。

1. **アスペクト比 ≈ 230:409**（スナップショットは正確に 690×1227 または 1380×2454）。
   16:9（1.77778）と非常に近いため、許容差は ±0.0003 と厳しめにして 16:9 を確実に除外しています。
2. **JPEG であること** — iOS の PNG スクリーンショットを除外。
3. **カメラ EXIF（Make/Model）が無いこと** — 似た比率にトリミングした実写真を除外。

precision（誤爆ゼロ）最優先の設計です。実写真を絶対にタグ付けしないことを目標にしています。
ピクセルはデコードせず、ヘッダと EXIF だけを読みます。

## インストール / ビルド

Go 1.23 以上が必要です。

```sh
go build -o photo-clean ./cmd/photo-clean
# または
make build
```

Nix を使う場合は後述の「開発環境（Nix + direnv）」を参照してください。

## 使い方

```sh
photo-clean scan  <dir>   # 走査して判定件数を表示（変更なし＝dry-run）
photo-clean tag   <dir>   # 判定が pokemon のファイルに pGo(赤)タグを付与（冪等）
photo-clean untag <dir>   # pGo(赤)タグだけを除去（やり直し・救済用）
```

- `-r` / `--recursive` でサブディレクトリも再帰的に走査します（デフォルトは非再帰）。
- `tag` は冪等で、既存タグを保持したまま `pGo` を追記します（2 回流しても二重に付きません）。
- `untag` は `pGo` タグだけを狙い撃ちで外し、ユーザーが付けた他のタグには触れません。

Makefile 経由でも実行できます。

```sh
make scan  DIR=~/Pictures REC=1
make tag   DIR=~/Pictures REC=1
make untag DIR=~/Pictures
make help    # ターゲット一覧
```

## 推奨ワークフロー

1. `scan` で件数と内訳を確認する。
2. 問題なさそうなら `tag` で `pGo`（赤）タグを付与する。
3. Finder で赤タグを絞り込み、目視確認しながら手動で削除する。
4. 付けすぎた場合は `untag` でやり直す。

## 開発環境（Nix + direnv）

`flake.nix` に Go ツール一式（go / gopls / gotools / staticcheck / delve）の dev shell を定義しています。

```sh
nix develop          # dev shell に入る
nix build .#default  # 再現ビルド（./result/bin/photo-clean）
```

direnv を使うと、このディレクトリに `cd` するだけで dev shell が自動で有効化されます。
初回のみ承認が必要です。

```sh
direnv allow   # 初回のみ
```

`.envrc` は nix-direnv の `use flake` を使っており、dev shell は `.direnv/` にキャッシュされるため
2 回目以降の `cd` は高速です。

## Finder タグの仕組み

Finder タグはファイルの拡張属性 `com.apple.metadata:_kMDItemUserTags` に、
`"名前\n色インデックス"` 形式の文字列配列をバイナリ plist で保存しています
（`pGo\n6`、6 = 赤）。`tag` は既存タグを保持して冪等に追記し、`untag` は `pGo` のみを除去します。

## ライセンス / 注意

- ローカル専用ツールです。ネットワークアクセスや外部コマンドへのシェルアウトはありません。
- macOS 専用です（Finder タグ＝拡張属性に依存）。
