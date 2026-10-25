package main

import (
	"io"
	"strings"
	"testing"
)

func TestDo(t *testing.T) {
	type expect struct {
		success     bool
		contentType string
	}

	tests := []struct {
		title        string
		text         string
		mode         string
		snippetMode  bool
		filepath     string
		deleteTarget MessageRef
		expect       expect
	}{
		{
			title:       "normal text",
			text:        "",
			mode:        modeText,
			snippetMode: false,
			expect: expect{
				success:     true,
				contentType: "message",
			},
		},
		{
			title:       "snippet mode",
			text:        strings.Repeat("a", 3000),
			mode:        modeText,
			snippetMode: true,
			expect: expect{
				success:     true,
				contentType: "file",
			},
		},
		{
			title:       "snippet (text over 3000 bytes)",
			text:        strings.Repeat("a", 3001),
			mode:        modeText,
			snippetMode: false,
			expect: expect{
				success:     true,
				contentType: "file",
			},
		},
		{
			title:       "upload file",
			snippetMode: true,
			mode:        modeFile,
			filepath:    "tests/upload.txt",
			expect: expect{
				success:     true,
				contentType: "file",
			},
		},
		{
			title:        "delete message",
			mode:         modeDelete,
			deleteTarget: MessageRef{channel: "C123", timestamp: "1700000000.000200"},
			expect: expect{
				success:     true,
				contentType: "delete",
			},
		},
		{
			title: "unknown mode",
			mode:  "",
			expect: expect{
				success:     false,
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			client := NewSlackMockClient()
			opts := &Options{
				slackClient: client,
				infoOut:     io.Discard,
				text:        tt.text,
				mode:        tt.mode,

				token:        "testtoken",
				filepath:     tt.filepath,
				snippetMode:  tt.snippetMode,
				deleteTarget: tt.deleteTarget,
				postOpts: &PostOptions{
					username:  "test username",
					channel:   "test channel",
					iconEmoji: "test icon emoji",
					iconUrl:   "test icon url",
				},
			}

			_, err := Do(opts)
			if err != nil && tt.expect.success {
				t.Errorf("error: Do %s", err.Error())
			}
			if err == nil && !tt.expect.success {
				t.Errorf("error: Do must be error (want:%v got:%v)", false, true)
			}

			if client.ContentType != tt.expect.contentType {
				t.Errorf("error: contentType (want:%v got:%v)", tt.expect.contentType, client.ContentType)
			}
		})
	}
}

func Test_parseMessageURL_validPermalink_returnsChannelAndTimestamp(t *testing.T) {
	tests := []struct {
		title string
		url   string
		want  MessageRef
	}{
		{"permalink", "https://example.slack.com/archives/C0123ABCD/p1234567890123456", MessageRef{channel: "C0123ABCD", timestamp: "1234567890.123456"}},
		{"thread reply permalink", "https://example.slack.com/archives/C0123ABCD/p1234567890123456?thread_ts=1234567890.000100&cid=C0123ABCD", MessageRef{channel: "C0123ABCD", timestamp: "1234567890.123456"}},
		{"DM permalink", "https://example.slack.com/archives/D0123ABCD/p1234567890123456", MessageRef{channel: "D0123ABCD", timestamp: "1234567890.123456"}},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got, err := parseMessageURL(tt.url)
			if err != nil {
				t.Fatalf("parseMessageURL(%q) error = %s", tt.url, err)
			}
			if got != tt.want {
				t.Errorf("parseMessageURL(%q) = %+v, want %+v", tt.url, got, tt.want)
			}
		})
	}
}

func Test_parseMessageURL_invalidURL_returnsError(t *testing.T) {
	tests := []struct {
		title string
		url   string
	}{
		{"empty", ""},
		{"unparsable url", "://example.slack.com/archives/C0123ABCD/p1234567890123456"},
		{"channel id only", "C0123ABCD"},
		{"channel url without message", "https://example.slack.com/archives/C0123ABCD"},
		{"missing p prefix", "https://example.slack.com/archives/C0123ABCD/1234567890123456"},
		{"not archives path", "https://app.slack.com/client/T0123ABCD/C0123ABCD"},
		{"too short timestamp", "https://example.slack.com/archives/C0123ABCD/p123456"},
		{"non digit timestamp", "https://example.slack.com/archives/C0123ABCD/p1234567890abcdef"},
		{"trailing path segment", "https://example.slack.com/archives/C0123ABCD/p1234567890123456/extra"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if got, err := parseMessageURL(tt.url); err == nil {
				t.Errorf("parseMessageURL(%q) = %+v, want error", tt.url, got)
			}
		})
	}
}

// strGetFirstOne は先頭から最初の非空文字列を返す。焼き込み値(embeddedToken /
// embeddedChannel)を末尾フォールバックとして使う優先順位の仕様を固定する。
func Test_strGetFirstOne_returnsFirstNonEmpty(t *testing.T) {
	tests := []struct {
		title string
		vars  []string
		want  string
	}{
		{"実行時指定が焼き込み値より優先", []string{"runtime", "", "embedded"}, "runtime"},
		{"先行が空なら次を採用", []string{"", "profile", "embedded"}, "profile"},
		{"全て空なら空文字", []string{"", "", ""}, ""},
		{"焼き込み値のみのフォールバック", []string{"", "", "embedded"}, "embedded"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if got := strGetFirstOne(tt.vars...); got != tt.want {
				t.Errorf("strGetFirstOne(%v) = %q, want %q", tt.vars, got, tt.want)
			}
		})
	}
}
