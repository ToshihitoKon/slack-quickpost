package main

import (
	"strings"
	"testing"
)

func TestDo(t *testing.T) {
	type expect struct {
		success     bool
		contentType string
	}

	tests := []struct {
		title       string
		text        string
		mode        string
		snippetMode bool
		filepath    string
		expect      expect
	}{
		{
			title:       "normal text",
			text:        "",
			mode:        "text",
			snippetMode: false,
			expect: expect{
				success:     true,
				contentType: "message",
			},
		},
		{
			title:       "snippet mode",
			text:        strings.Repeat("a", 3000),
			mode:        "text",
			snippetMode: true,
			expect: expect{
				success:     true,
				contentType: "file",
			},
		},
		{
			title:       "snippet (text over 3001 charactor)",
			text:        strings.Repeat("a", 3001),
			mode:        "text",
			snippetMode: false,
			expect: expect{
				success:     true,
				contentType: "file",
			},
		},
		{
			title:       "upload file",
			snippetMode: true,
			mode:        "file",
			filepath:    "tests/upload.txt",
			expect: expect{
				success:     true,
				contentType: "file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			client := NewSlackMockClient()
			opts := &Options{
				slackClient: client,
				text:        tt.text,
				mode:        tt.mode,

				token:       "testtoken",
				filepath:    tt.filepath,
				snippetMode: tt.snippetMode,
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
