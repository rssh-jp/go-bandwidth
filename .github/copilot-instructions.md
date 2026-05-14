# go-bandwidth — Copilot Instructions

## プロジェクト概要

`io.Reader` / `io.Writer` に対して帯域幅（スループット）制限を加えるライブラリ。
指定したバイト数／時間単位を超えないよう duty-cycle 方式でスリープを挟みながら Read/Write を実行する。

## パッケージ構成

```
go-bandwidth/
├── bandwidth.go       # ReadWriter 構造体・Functional Options・Read()/Write()
├── bandwidth_test.go  # テスト
└── cmd/bandwidth/
    └── main.go        # CLI エントリーポイント（動作確認用）
```

## コーディングルール

- Go 1.26 以上の構文・機能を使用する
- 外部依存ライブラリは持たない（標準ライブラリのみ）
- エラーは呼び出し元に返す。内部でのログ出力は禁止
- Functional Options パターン（`OptionXxx`）でパラメータを追加する
- `os.SEEK_*` は使用しない。`io.SeekStart` / `io.SeekCurrent` / `io.SeekEnd` を使う

## 設計上の注意点

### 帯域制限の仕組み（bandwidth.go）

- `constant` 構造体: `limit`（バイト上限）と `duration`（単位時間）を保持する
- `variable` 構造体: 現在の累積バイト数・タイマー・`sync.Mutex` を保持する
- `exec()` 関数: `taskCheck` → `taskExec` → `taskSleep` → `taskEnd` のステートマシンで制御
  - `bytes >= limit` になったら `taskSleep` に遷移し、残り時間スリープ後にカウンタをリセット
- `OptionUseDefault()` を使うと `defaultVariable` ポインタを共有する（複数インスタンス間で帯域を合算管理できる）

### デフォルト設定

- 初期値: `limit = 1024 * 1024`（1MiB）、`duration = 1s`
- `SetDefault()` で上書き可能
- `TestMain` でテスト用に上書きしているため、テストファイルの `SetDefault` 呼び出しは変更しないこと

### 新しい機能を追加する場合

- `bandwidth.go` に Functional Option を追加する
- `bandwidth_test.go` に対応するテストを追加する
- README.md の API リファレンステーブルも更新する

## テスト・検証コマンド

```bash
# ビルド確認
go build ./...

# テスト実行
go test ./...

# 静的解析
go vet ./...
```

## 依存更新

外部依存なし。標準ライブラリのみ使用のため `go mod tidy` のみで十分。

```bash
go mod tidy
```
