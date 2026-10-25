# PR のタイトルとラベル

- PR のタイトルは Conventional Commits の形式で書く
  - 例: `feat(cli): add --completion`、`fix: ...`、`docs: ...`
- 利用者に見える挙動を壊す変更には、タイトルの type / scope の直後に `!` を付ける
  - 例: `fix(cli)!: reject positional arguments`
  - 対象: stdout / stderr の出力の形、exit code、受け付けるフラグや引数、プロファイルや環境変数の読み方
  - `!` 付きのタイトルには `label-breaking-change` ワークフローが `breaking-change` ラベルを付け、リリースノートの「Breaking changes」の見出しにまとまる
- `!` を付け忘れて PR を作ったら、タイトルを直すか `breaking-change` ラベルを手で付ける
  - ワークフローはラベルを付けるだけで外さないため、誤って付いたラベルは手で外す
- dependabot の PR には触らない
  - `dependencies` ラベルが自動で付き、リリースノートの「Dependencies」の見出しにまとまる

# CHANGES.md

- 破壊的変更は CHANGES.md にも書く
  - 書くこと: 以前の挙動、新しい挙動、移行のしかた
  - 自動生成のリリースノートには PR のタイトルしか載らず、詳細はフッターのリンクから CHANGES.md を見てもらうため
- 次のリリースの版が決まっていなければ `## Unreleased` の節に書き、リリース時に版名へ直す

# リリース

- main に `vX.Y.Z` の注釈付きタグを打って push する
  - release ワークフロー（GoReleaser v2）が、`.github/release.yml` の見出しでリリースノートを自動生成する
