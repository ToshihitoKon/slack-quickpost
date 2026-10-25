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
- The home directory was previously looked up with `user.Current()` on
  every run, and a lookup failure was logged to stderr and always exited
  with code `1`, even without a profile. It is now looked up only when a
  profile is used (`--profile` / `SLACK_QUICKPOST_PROFILE`), taken from
  `$HOME` (`%USERPROFILE%` on Windows) instead of the user database, and a
  failure is reported via the `error` key like other failures, so
  `--nofail` applies to it.
- Specifying more than one of `--text`, `--textfile`, `--blocks` and
  `--file` is now an error, even when one of them has an empty value.
  Previously one of them was used silently, in the order `--text` >
  `--blocks` > `--textfile` > `--file`, and an empty value was skipped.
- Positional arguments were previously ignored silently. They are now an
  error.
- Success output previously had no trailing newline; it now does.
- `timestamp` previously appeared as `"timestamp":""` when unavailable; the
  key is now omitted entirely. As a result, `jq -r .timestamp` prints
  `null` instead of an empty string; use `jq -r '.timestamp // empty'` when
  passing the value to `--thread-ts`.
