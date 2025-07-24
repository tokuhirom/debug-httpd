# debug-httpd

デバッグ用途の HTTP サーバーです。環境変数やホスト情報を JSON 形式で返却します。

![screenshot](screenshot-20250724T181254@2x.png)

## 機能

このデバッグ用 HTTP サーバーは以下の情報を JSON 形式で返します：

- タイムスタンプ
- リクエスト情報（パス、ヘッダー、クライアントアドレス）
- ホスト情報（ホスト名、FQDN、IPアドレス）
- 環境変数
- Python バージョン

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
    "PYTHON_VERSION": "3.11.10",
    "HOME": "/root"
  },
  "python_version": "3.11.10 (main, Oct 16 2024, 02:31:39) [GCC 12.2.0]"
}
```

## CI/CD

GitHub Actions により、main ブランチへのプッシュ時に自動的に Docker イメージがビルドされ、GitHub Container Registry (ghcr.io) にプッシュされます。

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
