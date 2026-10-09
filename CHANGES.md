# Changes

## v1.0.0

### Breaking changes

- stdout previously mixed plain-text error messages and an `[INFO]` log
  line with the `CliOutput` JSON. It is now JSON-only; those messages moved
  to stderr and error messages are now conveyed via the `error` key
  (see [output](README.md#output)).
- Success output previously had no trailing newline; it now does.
- `timestamp` previously appeared as `"timestamp":""` when unavailable; the
  key is now omitted entirely.
