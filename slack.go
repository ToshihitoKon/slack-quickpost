package main

import (
	"github.com/slack-go/slack"
)

type SlackClient interface {
	UploadFile(slack.FileUploadParameters) (*slack.File, error)
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

type SlackClientMock struct {
	ContentType string
}

func NewSlackClientMock() *SlackClientMock {
	return &SlackClientMock{}
}

func (scm *SlackClientMock) UploadFile(_ slack.FileUploadParameters) (*slack.File, error) {
	scm.ContentType = "file"
	// TODO
	return nil, nil
}
func (scm *SlackClientMock) PostMessage(_ string, opts ...slack.MsgOption) (string, string, error) {
	scm.ContentType = "message"
	// TODO
	return "", "", nil
}
