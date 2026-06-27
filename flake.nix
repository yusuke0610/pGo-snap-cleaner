{
  # photo-clean — Pokémon GO の AR スナップショットに Finder タグ(pGo/赤)を付ける CLI
  description = "photo-clean: Pokémon GO スナップショット判定 & Finder タグ付与 CLI";

  inputs = {
    # 安定版が欲しければ nixos-25.05 などに固定してもよい
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      # 対応システム（Apple Silicon / Intel Mac / Linux）。
      # macOS 専用ツールだが、開発自体は Linux でもできるよう一応並べておく。
      systems = [ "aarch64-darwin" "x86_64-darwin" "x86_64-linux" "aarch64-linux" ];

      # 各システムについて関数 f を評価し { system = ...; } の属性集合を作るヘルパ
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f (import nixpkgs { inherit system; }));
    in
    {
      # 開発シェル: `nix develop`（direnv 経由なら cd するだけ）で Go ツール一式が入る
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          # シェルに入れる開発ツール
          packages = with pkgs; [
            go          # Go コンパイラ（go.mod の要求バージョン以上）
            gopls       # 言語サーバ（補完・定義ジャンプ）
            gotools     # goimports などの補助ツール
            go-tools    # staticcheck（静的解析）
            delve       # dlv デバッガ
          ];

          # シェル起動時に一度だけ走る。バージョンを表示するだけ。
          shellHook = ''
            echo "photo-clean dev shell — $(go version)"
          '';
        };
      });

      # `nix build` で単一バイナリをビルドする
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "photo-clean";
          version = "0.1.0";
          src = ./.;

          # ビルド対象のエントリポイント
          subPackages = [ "cmd/photo-clean" ];

          # 依存モジュールのハッシュ。go.mod / go.sum を変えたら
          # 一度 lib.fakeHash に戻してビルド→出力された正しい値に貼り替える。
          vendorHash = "sha256-bDeFPdErSp30SdsRQGNXyjhA+QrcEOIpsXbiKGnsojI=";

          meta = {
            description = "Pokémon GO スナップショットに Finder タグ(pGo/赤)を付ける macOS CLI";
            mainProgram = "photo-clean";
          };
        };
      });
    };
}
