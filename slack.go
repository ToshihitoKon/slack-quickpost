package main

import "github.com/slack-go/slack"

type SlackClient interface {
	UploadFile(slack.FileUploadParameters) (*slack.File, error)
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}
