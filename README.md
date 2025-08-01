# debug-httpd

デバッグ用途の軽量 HTTP サーバーです。Go言語で実装され、環境変数やホスト情報を JSON 形式で返却します。

![screenshot](screenshot-20250724T181254@2x.png)

## 機能

このデバッグ用 HTTP サーバーは以下の情報を JSON 形式で返します：

- タイムスタンプ
- リクエスト情報（パス、ヘッダー、クライアントアドレス）
- ホスト情報（ホスト名、FQDN、IPアドレス）
- 環境変数
- Go バージョン

## 使い方

### GitHub Container Registry から実行

```bash
# デフォルトポート（9876）で実行
docker run -p 9876:9876 ghcr.io/tokuhirom/debug-httpd:latest

# カスタムポートで実行（CMD経由）
docker run -p 8080:8080 ghcr.io/tokuhirom/debug-httpd:latest 8080

# カスタムポートで実行（環境変数経由）
docker run -p 8000:8000 -e PORT=8000 ghcr.io/tokuhirom/debug-httpd:latest
```

### ローカルでビルドして実行

```bash
# イメージをビルド
docker build -t debug-httpd:latest .

# コンテナを実行
docker run -p 9876:9876 debug-httpd:latest
```

### アクセス方法

```bash
# サーバーにアクセス
curl http://localhost:9876

# jq で整形して表示
curl -s http://localhost:9876 | jq .
```

## レスポンス例

```json
{
  "timestamp": "2024-01-01T12:00:00.123456",
  "request": {
    "path": "/",
    "headers": {
      "Host": "localhost:9876",
      "User-Agent": "curl/7.88.1",
      "Accept": "*/*"
    },
    "client_address": "172.17.0.1",
    "client_port": 54321
  },
  "host": {
    "hostname": "b2c58d189684",
    "fqdn": "b2c58d189684",
    "ip_addresses": ["172.17.0.2", "::1"]
  },
  "environment_variables": {
    "PATH": "/usr/local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "HOSTNAME": "b2c58d189684",
    "PORT": "9876",
    "HOME": "/root"
  },
  "go_version": "go1.21.0"
}
```

## エンドポイント

- `/` - デバッグ情報をJSONで返却
- `/ping` - ヘルスチェック用エンドポイント（"pong"を返す）
- `/logs` - 直近100件のアクセスログをJSONで返却

## ビルドとテスト

### ローカルでのビルド
```bash
# ユニットテストを実行
make test

# 統合テスト（Dockerが必要）を実行
make integration-test

# 全てのテストを実行
make test-all

# バイナリをビルド
make build

# 実行
make run
```

### Dockerイメージサイズ
約6.5MB (scratchベース) の超軽量イメージ

## CI/CD

GitHub Actions により：
- プルリクエスト時にテストとビルドを実行
- main ブランチへのプッシュ時に Docker イメージをビルドし、GitHub Container Registry (ghcr.io) にプッシュ

## リリース

### バージョニング

セマンティックバージョニング（例: v1.0.0）を使用します。

### リリース方法

```bash
# タグを作成してプッシュ
git tag v1.0.0
git push origin v1.0.0
```

タグをプッシュすると自動的に：
1. テスト（ユニット + 統合）を実行
2. マルチプラットフォーム向けバイナリをビルド
3. Docker イメージをビルドして ghcr.io にプッシュ
4. GitHub Release を作成

### 利用可能なDockerタグ

- `ghcr.io/tokuhirom/debug-httpd:latest` - 最新版
- `ghcr.io/tokuhirom/debug-httpd:v1.0.0` - 特定バージョン
- `ghcr.io/tokuhirom/debug-httpd:1.0` - マイナーバージョン
- `ghcr.io/tokuhirom/debug-httpd:1` - メジャーバージョン

## ライセンス

```
The MIT License (MIT)

Copyright © 2025 Tokuhiro Matsuno, http://64p.org/ <tokuhirom@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the “Software”), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
