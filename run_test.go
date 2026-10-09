package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func runWithMock(args []string) (exitCode int, stdout, stderr string) {
	var outBuf, errBuf bytes.Buffer
	newClient := func(token string) SlackClient {
		return NewSlackMockClient()
	}
	exitCode = runWith(args, &outBuf, &errBuf, newClient)
	return exitCode, outBuf.String(), errBuf.String()
}

func Test_runWith_postText(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--text", "hello"}
	code, stdout, _ := runWithMock(args)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "{\"channel\":\"C123\",\"timestamp\":\"1234567890.123456\"}\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func Test_runWith_postText_autoSnippetOverLimit(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	longText := strings.Repeat("a", 3001)
	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--text", longText}
	code, stdout, stderr := runWithMock(args)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	// postFileはtimestampを取得できないため、omitemptyによりtimestampキー自体が省略される。
	wantStdout := "{\"channel\":\"C123\"}\n"
	if stdout != wantStdout {
		t.Errorf("stdout = %q, want %q", stdout, wantStdout)
	}
	wantStderr := "[INFO] text length exceed limits (3000 characters), upload snippets.\n"
	if stderr != wantStderr {
		t.Errorf("stderr = %q, want %q", stderr, wantStderr)
	}
}

func Test_runWith_postText_explicitSnippet(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--text", "hello", "--snippet"}
	code, stdout, _ := runWithMock(args)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "{\"channel\":\"C123\"}\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func Test_runWith_postTextFile(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	f, err := os.CreateTemp(t.TempDir(), "textfile-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	if _, err := f.WriteString("hello from file"); err != nil {
		t.Fatalf("failed to write temp file: %s", err)
	}
	f.Close()

	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--textfile", f.Name()}
	code, stdout, _ := runWithMock(args)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "{\"channel\":\"C123\",\"timestamp\":\"1234567890.123456\"}\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func Test_runWith_postBlocks(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	// slack.Blocks.UnmarshalJSON requires a top-level array.
	blocksJSON := `[{"type":"section","text":{"type":"mrkdwn","text":"*Sample BlockKit*"}}]`
	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--blocks", blocksJSON}
	code, stdout, _ := runWithMock(args)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "{\"channel\":\"C123\",\"timestamp\":\"1234567890.123456\"}\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func Test_runWith_postFile(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--file", "tests/upload.txt"}
	code, stdout, _ := runWithMock(args)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "{\"channel\":\"C123\"}\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

type closedWriter struct{}

func (closedWriter) Write(p []byte) (int, error) {
	return 0, os.ErrClosed
}

func Test_runWith_stdoutWriteFailureAfterPost(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	mock := NewSlackMockClient()
	var errBuf bytes.Buffer
	args := []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--text", "hello"}
	code := runWith(args, closedWriter{}, &errBuf, func(token string) SlackClient {
		return mock
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if len(mock.PostMessageCalls) != 1 {
		t.Errorf("PostMessage calls = %d, want 1", len(mock.PostMessageCalls))
	}
	if !strings.Contains(errBuf.String(), "failed to write output") {
		t.Errorf("stderr = %q, want to contain \"failed to write output\"", errBuf.String())
	}
}

func Test_runWith_errorCases(t *testing.T) {
	tests := []struct {
		title      string
		args       []string
		wantSubstr string
	}{
		{
			title:      "token missing",
			args:       []string{"slack-quickpost", "--channel", "C123", "--text", "hello"},
			wantSubstr: "slack token is required",
		},
		{
			title:      "channel missing",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--text", "hello"},
			wantSubstr: "channel is required",
		},
		{
			title:      "no mode specified",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123"},
			wantSubstr: "--text option is required",
		},
		{
			title:      "invalid blocks json",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--blocks", "{not json"},
			wantSubstr: "failed to parse blocks json",
		},
		{
			title:      "textfile not found",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--textfile", "tests/not-exist.txt"},
			wantSubstr: "failed read text file",
		},
		{
			title:      "file not found",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--channel", "C123", "--file", "tests/not-exist.txt"},
			wantSubstr: "failed to open file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title+"/without nofail", func(t *testing.T) {
			t.Setenv("SLACK_TOKEN", "")
			t.Setenv("SLACK_QUICKPOST_PROFILE", "")

			code, stdout, stderr := runWithMock(tt.args)
			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			assertStdoutIsErrorJSON(t, stdout, tt.wantSubstr)
			if stderr != "" {
				t.Errorf("stderr = %q, want empty", stderr)
			}
		})

		t.Run(tt.title+"/with nofail", func(t *testing.T) {
			t.Setenv("SLACK_TOKEN", "")
			t.Setenv("SLACK_QUICKPOST_PROFILE", "")

			args := append(append([]string{}, tt.args...), "--nofail")
			code, stdout, stderr := runWithMock(args)
			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}
			assertStdoutIsErrorJSON(t, stdout, tt.wantSubstr)
			if stderr != "" {
				t.Errorf("stderr = %q, want empty", stderr)
			}
		})
	}
}

// assertStdoutIsErrorJSON はstdoutが「errorキーのみを持つJSON1行+改行」であり、
// そのerror文言がwantSubstrを含むことを検証する。ファイルパスやOS/Slackのエラー文言が
// 埋め込まれるケースがあるため、完全一致ではなく部分一致で照合する。
func assertStdoutIsErrorJSON(t *testing.T, stdout, wantSubstr string) {
	t.Helper()

	if !strings.HasSuffix(stdout, "\n") || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("stdout = %q, want exactly one line ending with newline", stdout)
	}

	// CliOutputにUnmarshalすると余分なキーを検出できないため、mapで受けてキー数を見る
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %q: %s", stdout, err)
	}
	if len(got) != 1 {
		t.Errorf("stdout = %q, want only the error key", stdout)
	}
	errMsg, ok := got["error"].(string)
	if !ok || !strings.Contains(errMsg, wantSubstr) {
		t.Errorf("error = %v, want a string containing %q", got["error"], wantSubstr)
	}
}

func Test_runWith_unknownFlag(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	tests := []struct {
		title string
		args  []string
	}{
		{"without nofail", []string{"slack-quickpost", "--foo"}},
		{"with nofail", []string{"slack-quickpost", "--foo", "--nofail"}},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// An unknown flag always exits with 2, regardless of --nofail.
			code, stdout, stderr := runWithMock(tt.args)
			if code != 2 {
				t.Errorf("exit code = %d, want 2", code)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			// main.go manually reproduces pflag's ExitOnError output order
			// ("Usage of ...:" -> flag list -> "unknown flag: ..."); verify it holds.
			usageIdx := strings.Index(stderr, "Usage of")
			nofailIdx := strings.Index(stderr, "--nofail")
			unknownIdx := strings.Index(stderr, "unknown flag: --foo")
			if usageIdx == -1 || nofailIdx == -1 || unknownIdx == -1 {
				t.Fatalf("stderr = %q, want to contain \"Usage of\", \"--nofail\" and \"unknown flag: --foo\"", stderr)
			}
			if !(usageIdx < nofailIdx && nofailIdx < unknownIdx) {
				t.Errorf("stderr order wrong: want Usage -> flag list -> unknown flag, got %q", stderr)
			}
		})
	}
}

func Test_runWith_version(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	code, stdout, _ := runWithMock([]string{"slack-quickpost", "--version"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	// Output depends on buildinfo, so only check that exactly one line is printed.
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 1 || lines[0] == "" {
		t.Errorf("stdout = %q, want exactly one non-empty line", stdout)
	}
}

func Test_runWith_help(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	tests := []struct {
		title string
		args  []string
	}{
		{"--help", []string{"slack-quickpost", "--help"}},
		{"-h", []string{"slack-quickpost", "-h"}},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Usage text depends on the program name, so stderr is not checked.
			code, stdout, _ := runWithMock(tt.args)
			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
		})
	}
}
