# go-bandwidth

帯域幅（スループット）を制限しながら `io.Reader` / `io.Writer` を扱うためのライブラリです。
指定したバイト数／時間単位を超えないよう duty-cycle 方式でスリープを挟みながら Read/Write を実行します。

## 特徴

- `io.Reader` / `io.Writer` / `io.ReadWriter` すべてに対応
- Functional Options パターンによる柔軟な設定
- デフォルト設定の共有インスタンス（`OptionUseDefault`）をサポート
- 外部依存なし（標準ライブラリのみ）

## インストール

```bash
go get github.com/rssh-jp/go-bandwidth
```

## 使い方

### Writer（書き込み帯域制限）

```go
package main

import (
    "log"
    "os"
    "time"

    bandwidth "github.com/rssh-jp/go-bandwidth"
)

func main() {
    fd, err := os.OpenFile("./output", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer fd.Close()

    // 1秒あたり 1024 バイトに帯域制限して書き込む
    w := bandwidth.New(
        bandwidth.OptionWriter(fd),
        bandwidth.OptionConstant(1024, time.Second),
    )

    for i := 0; i < 100; i++ {
        if _, err := w.Write([]byte{byte(i)}); err != nil {
            log.Fatal(err)
        }
    }
}
```

### Reader（読み込み帯域制限）

```go
r := bandwidth.New(
    bandwidth.OptionReader(fd),
    bandwidth.OptionConstant(1024, time.Second),
)

buf := make([]byte, 4096)
n, err := r.Read(buf)
```

### ReadWriter（読み書き両方）

```go
rw := bandwidth.New(
    bandwidth.OptionReader(fd),
    bandwidth.OptionWriter(fd),
    bandwidth.OptionConstant(1024, time.Second),
)
```

### デフォルト設定の利用

```go
// デフォルト設定（初期値: 1MiB/秒）を変更する
bandwidth.SetDefault(512*1024, time.Second)

// デフォルト設定を共有するインスタンスを作成
w := bandwidth.New(bandwidth.OptionUseDefault(), bandwidth.OptionWriter(fd))
```

## API リファレンス

### 関数

| 関数 | 説明 |
|------|------|
| `New(options ...Option) *ReadWriter` | `ReadWriter` インスタンスを生成する |
| `SetDefault(limit int64, duration time.Duration)` | デフォルトの帯域上限を設定する |

### Functional Options

| オプション | 説明 |
|-----------|------|
| `OptionReader(r io.Reader) Option` | 帯域制限付きリーダーを設定する |
| `OptionWriter(w io.Writer) Option` | 帯域制限付きライターを設定する |
| `OptionConstant(limit int64, duration time.Duration) Option` | 帯域上限を個別に設定する |
| `OptionUseDefault() Option` | グローバルデフォルト設定・変数を共有する |

### エラー

| 変数 | 説明 |
|------|------|
| `ErrCouldNotFoundFunction` | Reader/Writer が設定されていない状態で Read/Write を呼んだ場合に返る |

## テスト・ビルド

```bash
# ビルド確認
go build ./...

# テスト実行
go test ./...

# 静的解析
go vet ./...
```

## 依存

外部依存なし。標準ライブラリのみ使用。

```bash
go mod tidy
```
