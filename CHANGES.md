# Changes

## v1.0.0

### Breaking changes

- stdout previously mixed plain-text error messages and an `[INFO]` log
  line with the `CliOutput` JSON. It is now JSON-only: error messages are
  conveyed via the `error` key (see [output](README.md#output)), and the
  `[INFO]` line moved to stderr.
- Error messages no longer start with `error:`, and several were reworded
  (e.g. `error open file` → `failed to open file`,
  `error: failed blocks.UnmarshalJSON` → `failed to parse blocks json`).
  Detect failures by the exit code or the presence of the `error` key
  instead of matching message text.
- A `user.Current()` failure was previously logged to stderr and always
  exited with code `1`. It is now reported via the `error` key like other
  failures, so `--nofail` applies to it.
- Success output previously had no trailing newline; it now does.
- `timestamp` previously appeared as `"timestamp":""` when unavailable; the
  key is now omitted entirely. As a result, `jq -r .timestamp` prints
  `null` instead of an empty string; use `jq -r '.timestamp // empty'` when
  passing the value to `--thread-ts`.
