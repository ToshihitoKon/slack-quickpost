package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func runWithMock(args []string) (exitCode int, stdout, stderr string) {
	return runWithClient(args, NewSlackMockClient())
}

func runWithClient(args []string, client *SlackMockClient) (exitCode int, stdout, stderr string) {
	var outBuf, errBuf bytes.Buffer
	newClient := func(token string) SlackClient {
		return client
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
	wantStderr := "[INFO] text length exceed limits (3000 bytes), upload snippets.\n"
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

func Test_runWith_stdoutWriteFailureAfterDelete(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	mock := NewSlackMockClient()
	var errBuf bytes.Buffer
	args := []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1700000000000200"}
	code := runWith(args, closedWriter{}, &errBuf, func(token string) SlackClient {
		return mock
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if len(mock.DeleteMessageCalls) != 1 {
		t.Errorf("DeleteMessage calls = %d, want 1", len(mock.DeleteMessageCalls))
	}
	if !strings.Contains(errBuf.String(), "failed to write output") {
		t.Errorf("stderr = %q, want to contain \"failed to write output\"", errBuf.String())
	}
}

// 削除対象の ts は mockPostMessageTimestamp と別の値にし、誤って PostMessage が呼ばれた場合と区別できるようにする
func Test_runWith_deleteMessage(t *testing.T) {
	tests := []struct {
		title      string
		url        string
		extraArgs  []string
		wantTarget MessageRef
	}{
		{"permalink", "https://example.slack.com/archives/C123/p1700000000000200", nil, MessageRef{channel: "C123", timestamp: "1700000000.000200"}},
		// スレッド返信の permalink。thread_ts は親の ts なので、削除対象は p 以降の ts になる
		{"thread reply permalink", "https://example.slack.com/archives/C123/p1700000000000200?thread_ts=1700000000.000100&cid=C123", nil, MessageRef{channel: "C123", timestamp: "1700000000.000200"}},
		{"with nofail", "https://example.slack.com/archives/C123/p1700000000000200", []string{"--nofail"}, MessageRef{channel: "C123", timestamp: "1700000000.000200"}},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			t.Setenv("SLACK_TOKEN", "")
			t.Setenv("SLACK_QUICKPOST_PROFILE", "")

			client := NewSlackMockClient()
			args := append([]string{"slack-quickpost", "--token", "xoxb-test", "--delete", tt.url}, tt.extraArgs...)
			code, stdout, _ := runWithClient(args, client)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0", code)
			}
			if len(client.DeleteMessageCalls) != 1 || client.DeleteMessageCalls[0] != tt.wantTarget {
				t.Errorf("DeleteMessage calls = %+v, want [%+v]", client.DeleteMessageCalls, tt.wantTarget)
			}
			if len(client.PostMessageCalls) != 0 || len(client.UploadFileCalls) != 0 {
				t.Errorf("unexpected post calls: PostMessage=%v UploadFile=%d", client.PostMessageCalls, len(client.UploadFileCalls))
			}
			want := "{\"channel\":\"C123\",\"timestamp\":\"1700000000.000200\"}\n"
			if stdout != want {
				t.Errorf("stdout = %q, want %q", stdout, want)
			}
		})
	}
}

func Test_runWith_deleteMessage_ignoresDefaultChannel(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")
	original := embeddedChannel
	embeddedChannel = "CDEFAULT"
	t.Cleanup(func() { embeddedChannel = original })

	client := NewSlackMockClient()
	args := []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1700000000000200"}
	code, _, _ := runWithClient(args, client)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := MessageRef{channel: "C123", timestamp: "1700000000.000200"}
	if len(client.DeleteMessageCalls) != 1 || client.DeleteMessageCalls[0] != want {
		t.Errorf("DeleteMessage calls = %+v, want [%+v]", client.DeleteMessageCalls, want)
	}
}

func Test_runWith_deleteMessage_rejectedInput_doesNotDelete(t *testing.T) {
	url := "https://example.slack.com/archives/C123/p1700000000000200"
	tests := []struct {
		title    string
		args     []string
		wantCode int
	}{
		{"version takes precedence", []string{"slack-quickpost", "--token", "xoxb-test", "--delete", url, "--version"}, 0},
		{"help takes precedence", []string{"slack-quickpost", "--token", "xoxb-test", "--delete", url, "--help"}, 0},
		{"completion takes precedence", []string{"slack-quickpost", "--token", "xoxb-test", "--delete", url, "--completion", "zsh"}, 0},
		{"conflicting option", []string{"slack-quickpost", "--token", "xoxb-test", "--delete", url, "--text", "hello"}, 1},
		{"extra url", []string{"slack-quickpost", "--token", "xoxb-test", "--delete", url, url}, 1},
		{"repeated delete", []string{"slack-quickpost", "--token", "xoxb-test", "--delete", url, "--delete", url}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			t.Setenv("SLACK_TOKEN", "")
			t.Setenv("SLACK_QUICKPOST_PROFILE", "")

			client := NewSlackMockClient()
			code, _, _ := runWithClient(tt.args, client)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if len(client.DeleteMessageCalls) != 0 {
				t.Errorf("DeleteMessage calls = %+v, want none", client.DeleteMessageCalls)
			}
		})
	}
}

func Test_runWith_deleteMessage_apiError(t *testing.T) {
	tests := []struct {
		title     string
		extraArgs []string
		wantCode  int
	}{
		{"without nofail", nil, 1},
		{"with nofail", []string{"--nofail"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			t.Setenv("SLACK_TOKEN", "")
			t.Setenv("SLACK_QUICKPOST_PROFILE", "")

			client := &SlackMockClient{DeleteMessageErr: errors.New("cant_delete_message")}
			args := append([]string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1700000000000200"}, tt.extraArgs...)
			code, stdout, _ := runWithClient(args, client)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			assertStdoutIsErrorJSON(t, stdout, "failed to delete message: cant_delete_message")
		})
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
		{
			title:      "delete with post mode option",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1234567890123456", "--text", "hello"},
			wantSubstr: "--delete cannot be used with --text",
		},
		{
			title:      "delete with channel",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1234567890123456", "--channel", "C123"},
			wantSubstr: "--delete cannot be used with --channel",
		},
		{
			// 空文字指定でも排他違反として扱う。複数あれば flag 名順にすべて列挙する
			title:      "delete with multiple options including empty value",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1234567890123456", "--username", "", "--snippet"},
			wantSubstr: "--delete cannot be used with --snippet, --username",
		},
		{
			title:      "delete with extra url",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1234567890123456", "https://example.slack.com/archives/C123/p1234567890999999"},
			wantSubstr: "--delete takes exactly one url, got extra arguments: https://example.slack.com/archives/C123/p1234567890999999",
		},
		{
			title:      "delete specified twice",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123/p1234567890123456", "--delete=https://example.slack.com/archives/C123/p1234567890999999"},
			wantSubstr: "--delete can be specified only once",
		},
		{
			title:      "delete with invalid url",
			args:       []string{"slack-quickpost", "--token", "xoxb-test", "--delete", "https://example.slack.com/archives/C123"},
			wantSubstr: "invalid message url (want a permalink like https://<workspace>.slack.com/archives/<channel>/p<timestamp>)",
		},
		{
			title:      "delete token missing",
			args:       []string{"slack-quickpost", "--delete", "https://example.slack.com/archives/C123/p1234567890123456"},
			wantSubstr: "slack token is required",
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

func Test_runWith_completion(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	for shell, script := range completionScripts {
		t.Run(shell, func(t *testing.T) {
			code, stdout, stderr := runWithMock([]string{"slack-quickpost", "--completion", shell})
			if code != 0 {
				t.Fatalf("exit code = %d, want 0", code)
			}
			if stdout != script {
				t.Errorf("stdout is not the %s completion script: %q", shell, stdout)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want empty", stderr)
			}
		})
	}
}

func Test_runWith_completion_unsupportedShell_isFlagParseError(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	code, stdout, stderr := runWithMock([]string{"slack-quickpost", "--completion", "fish", "--nofail"})
	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "unsupported shell (want bash or zsh)") {
		t.Errorf("stderr = %q, want to contain \"unsupported shell (want bash or zsh)\"", stderr)
	}
}

func matchedFlags(pattern, s string) []string {
	var flags []string
	for _, m := range regexp.MustCompile(pattern).FindAllStringSubmatch(s, -1) {
		flags = append(flags, m[1])
	}
	slices.Sort(flags)
	return slices.Compact(flags)
}

func Test_completionScripts_listSameFlagsAsUsage(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "")
	t.Setenv("SLACK_QUICKPOST_PROFILE", "")

	_, _, usage := runWithMock([]string{"slack-quickpost", "--help"})
	want := matchedFlags(`(?m)^\s+(?:-\w, )?(--[\w-]+)`, usage)
	if len(want) == 0 {
		t.Fatalf("no flags found in usage: %q", usage)
	}

	bashFlagList := regexp.MustCompile(`compgen -W "--[^"]*"`).FindString(bashCompletion)
	// pflag が暗黙に足す --help は usage に出ないため比較から外す。zsh は {-h,--help} と書くのでパターンに掛からない
	got := map[string][]string{
		"zsh":  matchedFlags(`(--[\w-]+)=?\[`, zshCompletion),
		"bash": slices.DeleteFunc(matchedFlags(`(--[\w-]+)`, bashFlagList), func(f string) bool { return f == "--help" }),
	}
	for shell, flags := range got {
		if !slices.Equal(flags, want) {
			t.Errorf("%s completion flags = %v, want %v", shell, flags, want)
		}
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
