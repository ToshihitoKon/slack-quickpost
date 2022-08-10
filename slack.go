package main

import "github.com/slack-go/slack"

type SlackClient interface {
	UploadFile(slack.FileUploadParameters) (*slack.File, error)
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

type SlackClientMock struct{}

func NewSlackClientMock() *SlackClientMock {
	return &SlackClientMock{}
}

func (scm *SlackClientMock) UploadFile(params slack.FileUploadParameters) (*slack.File, error) {
	// TODO
	return nil, nil
}

func (scm *SlackClientMock) PostMessage(_ string, _ ...slack.MsgOption) (string, string, error) {
	// TODO
	return "", "", nil
}
