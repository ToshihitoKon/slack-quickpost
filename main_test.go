package main

import "testing"

func TestDo(t *testing.T) {
	client := NewSlackClientMock()
	opts := &Options{
		slackClient: client,
		token:       "testtoken",
		text:        "test text",
		filepath:    "test filepath",
		mode:        "text",
		snippetMode: false,
		postOpts: &PostOptions{
			username:  "test username",
			channel:   "test channel",
			iconEmoji: "test icon emoji",
			iconUrl:   "test icon url",
		},
	}
	if err := Do(opts); err != nil {
		t.Errorf("error: Do %s", err.Error())
	}
}
