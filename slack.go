package main

import (
	"github.com/slack-go/slack"
)

type SlackClient interface {
	UploadFileV2(slack.UploadFileV2Parameters) (*slack.FileSummary, error)
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

type SlackMockClient struct {
	ContentType string
}

func NewSlackMockClient() *SlackMockClient {
	return &SlackMockClient{}
}

func (smc *SlackMockClient) UploadFileV2(_ slack.UploadFileV2Parameters) (*slack.FileSummary, error) {
	smc.ContentType = "file"
	// TODO
	return nil, nil
}

func (smc *SlackMockClient) PostMessage(_ string, opts ...slack.MsgOption) (string, string, error) {
	smc.ContentType = "message"
	// TODO
	return "", "", nil
}
