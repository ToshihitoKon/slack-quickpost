package main

import (
	"github.com/slack-go/slack"
)

type SlackClient interface {
	UploadFile(slack.UploadFileParameters) (*slack.FileSummary, error)
	PostMessage(string, ...slack.MsgOption) (string, string, error)
	DeleteMessage(channel, messageTimestamp string) (string, string, error)
	DeleteFile(fileID string) error
}
