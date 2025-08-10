# laminate: image generator bridge

## 概要

- このツールは文字列(主にコード文字列)を画像に変換するためのツールです
- 標準入力から文字列を読み取り、それを画像変換した結果を標準出力に出力します
    - 入力
        - コード文字列: 標準入力
        - コードの言語(ex. QRコード、Mermaid、その他のプログラミング言語) : `CODEBLOCK_LANG` 環境変数 or `--lang` flag (Optional)
    - 出力
        - 生成された画像: 標準出力
- langの値に応じた画像生成コマンドを実行します
    - 画像生成コマンドはYAML設定ファイルで指定します
- キャッシュ機構を備えます


## 設定ファイル

### 例

```yaml
# example
cache: 1h
commands:
- lang: qr
  run: 'qrencode -o "{{output}}" -t png "{{input}}"'
- lang: mermaid
  run: 'mmdc -i - -o "{{output}}" --quiet'
- lang: '{c,cpp,python,rust,go,java,js,ts,html,css,sh}'
  run: 'silicon -l "{{lang}}" -o "{{output}}"'
  ext: png
- lang: '*'
  run: [convert, -background, none, -fill, black, label:{{input}}, {{output}}]
  ext: png
```

### スキーマ

```yaml
# schema
type: object
properties:
  cache:
    type: string
    description: |
      キャッシュの有効期限を指定します。
      例: `1h`, `30m`, `15s` など。省略時はキャッシュは無効になります。
      キャッシュは、入力文字列とlangに基づいて生成されます。
      キャッシュが有効な場合、コマンドは実行されず、キャッシュから画像が読み込まれます。
  commands:
    type: array
    items:
      type: object
      properties:
        lang:
          type: string
          description: |
            言語を指定します。
            zsh的なglobパターンで指定できます。例: `{c,cpp,python,rust,go,java,js,ts,html,css,sh}`, `*`
        run:
          type: [string, array]
          items:
            type: string
          description: |
            コマンドラインで実行されるコマンドを指定します。
            {{input}}、{{output}}、{{lang}} 変数が展開されたあとに実行されます。
            文字列で指定された場合、 `{shell} -c` で実行されます。
            配列で指定された場合、そのまま実行されます。
        ext:
          type: string
          description: '出力画像の拡張子を指定します。省略時は png が使用されます。'
          pattern: '^[a-zA-Z0-9]+(?:\.[a-zA-Z0-9]+)*$'
        shell:
          type: string
          description: |
            コマンドを実行するシェルを指定します。
            省略時はbash、bashが見つからない場合は sh が使用されます。
      required:
        - lang
        - run

```

- 実行コマンドの決定
    - `commands:` 配列の先頭から走査し、最初にマッチしたlangに対応するコマンドが実行されます
- 変数展開に関するルール
    - `{{input}}`: `laminate` が標準入力から読み取ったコード文字列
        - この変数がない場合は、コマンドに対して標準入力経由で文字列が渡されます
    - `{{output}}`: 生成される画像の出力先
        - 標準では `.png` 拡張子が使用されますが、`ext` が指定されている場合はその拡張子が使用されます
        - この変数がない場合は、コマンドの標準出力を画像データとして読み取ります
    - `{{lang}}`: コマンド実行時に指定された言語を表します

## 利用ライブラリ
- github.com/gobwas/glob
    - lang指定の文字列マッチのために利用
- github.com/goccy/go-yaml
    - YAML設定ファイルの読み書きのために利用。YAMLライブラリは色々ありますがこれを使います
- github.com/k1LoW/exec
    - `os/exec` を使う代わりに一律このライブラリを使います
- github.com/spf13/pathologize
    - ファイルネームのサニタイズのために利用します

## 設定ファイルやキャッシュの位置
- XDG Base Directory に準じます
    - macであっても同様です
        - これは、 `os.UserConfigDir(), `os.UserCacheDir()`等を「使わない」と言うことです
            - これらはmacの場合 `~/Library/Application Support` などを返すため
            - デスクトップアプリケーションであればそれで良いが、CLIツールでは `XDG` に寄せた方が分かりやすいと考えます
    - windowsでは、 `os.UserConfigDir()` 及び `os.UserCacheDir()` を使用します
- 設定ファイルは `${XDG\_CONFIG\_HOME:-~/.config}/laminate/config.yaml` に配置されます
- キャッシュは `${XDG\_CACHE\_HOME:-~/.cache}/laminate/cache` に配置されます

## キャッシュキー
- キャッシュファイル名は、入力文字列とlang、拡張子に基づいて以下のように導出されます
    - `{{lang}}/{{hash(input)}}.{{ext}}`
        - 例: `qr/5d41402abc4b2a76b9719d911017c592.png`
        - hashはmd5を使用します
        - langはファイルシステムセーフな文字列にサニタイズされます
- キャッシュファイルの内容は、画像データそのものです
- キャッシュファイルのmtimeを見て有効期限の判定を行います

## 参考
- <https://github.com/haya14busa/cachecmd>
- <https://github.com/k1LoW/deck>
