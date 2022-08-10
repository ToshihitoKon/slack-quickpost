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
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			client := NewSlackClientMock()
			opts := &Options{
				slackClient: client,
				text:        tt.text,
				mode:        tt.mode,

				token:       "testtoken",
				filepath:    "test filepath",
				snippetMode: tt.snippetMode,
				postOpts: &PostOptions{
					username:  "test username",
					channel:   "test channel",
					iconEmoji: "test icon emoji",
					iconUrl:   "test icon url",
				},
			}

			err := Do(opts)
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
