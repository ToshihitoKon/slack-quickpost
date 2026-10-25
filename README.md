# slack-quickpost

## installation

```
go install github.com/ToshihitoKon/slack-quickpost@latest
```

## setup

https://api.slack.com/apps/  

make Slack App and get OAuth token.

### GitHub Actions

```yaml
jobs:
  using-slack-quickpost:
    runs-on: ubuntu-latest
    steps:
      - uses: ToshihitoKon/slack-quickpost@v1
        with:
          version: 0.10.0

```

## usage

OAuth token given by one of the following methods.

#### environment variable

```
export SLACK_TOKEN="xoxb-XXXXXXXX-XXXXXXX-XXXXXX"
slack-quickpost \
  --channel [CHANNEL_ID] \
  --text [TEXT]
```

#### Option

```
slack-quickpost \
  --token xoxb-XXXXXXXX-XXXXXXX-XXXXXX \
  --channel [CHANNEL_ID] \
  --text [TEXT]
```

#### Config file

Save config yaml `~/.config/slack-quickpost/profile-name.yaml`

```yaml
token: xoxb-XXX
channel: XXX
```

Provide the profile name using an environment variable(`SLACK_QUICKPOST_PROFILE`) or option.

```
SLACK_QUICKPOST_PROFILE=profile-name slack-quickpost --text [TEXT]
slack-quickpost --profile profile-name --text [TEXT]
```

The `.yaml` extension may be included in the profile name
(e.g. `--profile profile-name.yaml`).

#### Precedence

When values come from several sources, the first one found is used:

- token: `--token` → `SLACK_TOKEN` → profile → embedded value
- channel: `--channel` → profile → embedded value

A profile does not override `SLACK_TOKEN`: with `SLACK_TOKEN` exported,
`--profile` only supplies the channel.

#### Embedded binary

Build a `slack-quickpost` binary with a token and default channel baked in, so
that end users can run it without providing any credentials.

Install the builder:

```
go install github.com/ToshihitoKon/slack-quickpost/cmd/slack-quickpost-embed@latest
```

Build a binary with values embedded:

```
# fetch the published module and embed values (default)
slack-quickpost-embed \
  --token xoxb-XXXXXXXX-XXXXXXX-XXXXXX \
  --channel [CHANNEL_ID] \
  --output ./slack-quickpost

# token / channel can also come from environment variables
SLACK_TOKEN=xoxb-XXX SLACK_CHANNEL=[CHANNEL_ID] slack-quickpost-embed --output ./slack-quickpost

# pin the version to embed (v0.9.1 or later)
slack-quickpost-embed --token xoxb-XXX --channel [CHANNEL_ID] --version v0.10.0

# build from the current source tree instead of the published module
slack-quickpost-embed --token xoxb-XXX --channel [CHANNEL_ID] --source local

# cross compile for linux/arm64
slack-quickpost-embed \
  --token xoxb-XXX --channel [CHANNEL_ID] \
  --goos linux --goarch arm64 \
  --output ./slack-quickpost-linux-arm64

# cross compile for windows/amd64 (".exe" is appended automatically)
slack-quickpost-embed \
  --token xoxb-XXX --channel [CHANNEL_ID] \
  --goos windows --goarch amd64 \
  --output ./slack-quickpost
# -> ./slack-quickpost.exe
```

The resulting binary uses the embedded token / channel only as a fallback;
runtime inputs still take precedence (see [Precedence](#precedence)).

- Use v0.9.1 or later with `--version`. Versions before v0.9.0 have no
  embedding support: the build still succeeds, but nothing is embedded.
  v0.9.0 is retracted.
- The token is stored in the binary as plain text. Anyone who has the
  binary can extract it and use it with all of its scopes, so embed a
  token with only the scopes you intend to share.

Options:

```
--token string     slack app OAuth token to embed (fallback: SLACK_TOKEN)
--channel string   default slack channel id to embed (fallback: SLACK_CHANNEL)
--output string    output binary path (default "slack-quickpost")
--source string    build source: "module" (fetch by version) or "local" (current directory) (default "module")
--version string   slack-quickpost version to build when --source=module (default "latest")
--goos string      target GOOS for cross compilation (default: host GOOS)
--goarch string    target GOARCH for cross compilation (default: host GOARCH)
```

### post text

```
slack-quickpost \
  --channel [CHANNEL_ID] \
  --text [TEXT] \
  --username [DISPLAY_USERNAME] \
  --icon [ICON_EMOJI]

# text given textfile path and icon given image url
slack-quickpost \
  --channel [CHANNEL_ID] \
  --textfile [FILEPATH] \
  --username [DISPLAY_USERNAME] \
  --icon-url [ICON_IMAGE_URL]

# post text as snippet
slack-quickpost \
  --channel [CHANNEL_ID] \
  --text [TEXT] \
  --snippet

# post BlockKit
slack-quickpost \
  --channel [CHANNEL_ID] \
  --blocks '[{"type":"section","text":{"type":"mrkdwn","text":"*Sample BlockKit*"}}]'
```

In `--text`, the two characters `\n` are converted to a newline.
`--textfile` is posted as is.

`--blocks` takes a JSON array of blocks, not a single block object.

### post file

```
slack-quickpost \
  --channel [CHANNEL_ID] \
  --file [FILE_PATH]
```

### delete message

Pass a message permalink (the URL from "Copy link" in Slack). The channel and
timestamp are taken from the URL. Quote the URL: permalinks of thread replies
contain `?` and `&`.

```
slack-quickpost \
  --delete 'https://[WORKSPACE].slack.com/archives/[CHANNEL_ID]/p[TIMESTAMP]'
```

- `--delete` can only be combined with `--token`, `--profile` and `--nofail`.
  Any other option (including `--channel`), a second `--delete` or an extra
  positional argument is an error. `SLACK_TOKEN` / `SLACK_QUICKPOST_PROFILE`
  work as usual.
- `--version` and `--help` still take precedence and exit without deleting
  anything.
- The default channel from a profile or an embedded binary is ignored.
- The output has the same shape as a post, with the deleted message's
  `channel` and `timestamp`.
- With a bot token, Slack only allows deleting messages posted by that bot
  (`chat.delete` API restriction).
- File and snippet posts (including auto-snippets over 3000 bytes) cannot be
  deleted: Slack gives them only a file link (`/files/...`), not a message
  permalink.
- Embedded binaries can use `--delete` too, so anyone who has the binary can
  delete whatever its embedded token is allowed to delete.

### shell completion

`--completion` prints a completion script for bash or zsh. `--profile`
completes the profile names in `~/.config/slack-quickpost/`.

```
# zsh: add to ~/.zshrc after compinit
eval "$(slack-quickpost --completion zsh)"

# bash: add to ~/.bashrc
eval "$(slack-quickpost --completion bash)"
```

For zsh, you can instead save the output of `slack-quickpost --completion zsh`
as `_slack-quickpost` in a directory on `$fpath`.

## output

stdout always prints exactly one line of `CliOutput` JSON, followed by a
trailing newline.

On success:

```json
{"channel":"C123","timestamp":"1234567890.123456"}
```

The `timestamp` key is omitted when it is not available (e.g. `--file` /
snippet uploads, since the Slack API used for file uploads does not return a
message timestamp).

For the same uploads, `channel` is the requested channel value, not a value
echoed back from the Slack API — the upload API response does not include
one.

On failure, only the `error` key is present, regardless of `--nofail`:

```json
{"error":"slack token is required"}
```

`--nofail` only changes the exit code (`1` → `0` on failure); it never
changes what is printed to stdout. Human-readable diagnostics (e.g. the
`[INFO]` notice printed when text is auto-uploaded as a snippet for
exceeding 3000 bytes) go to stderr, never stdout.

Exceptions to "stdout is JSON-only":
- `--version` prints a plain version string to stdout.
- `--help` / `-h` prints the usage text to stderr and exits with code `0`;
  nothing is printed to stdout.
- A flag parse error (an unknown flag, a flag missing its value, an invalid
  boolean value, etc.) prints pflag's usage text to stderr and exits with
  code `2`; nothing is printed to stdout, and `--nofail` has no effect on
  this exit code.

See [CHANGES.md](CHANGES.md) for breaking changes in v1.0.

## command options

```
--blocks string      post BlockKit json
--channel string     post slack channel id
--completion string  print shell completion script (bash or zsh)
--delete string      delete message by permalink url
--file string        post file path
--icon string        icon emoji
--icon-url string    icon image url
--nofail             exit with code 0 even on failure (except flag parse errors)
--profile string     slack quickpost profile name
--snippet            post text as snippet
--text string        post text
--textfile string    post text file path
--thread-ts string   post under thread
--token string       slack app OAuth token
--username string    user name
--version            print version
```
