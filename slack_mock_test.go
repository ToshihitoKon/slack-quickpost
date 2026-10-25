package main

import (
	"github.com/slack-go/slack"
)

const mockPostMessageTimestamp = "1234567890.123456"

type SlackMockClient struct {
	ContentType string

	UploadFileCalls    []slack.UploadFileParameters
	PostMessageCalls   []string
	DeleteMessageCalls []MessageRef

	DeleteMessageErr error
}

func NewSlackMockClient() *SlackMockClient {
	return &SlackMockClient{}
}

func (smc *SlackMockClient) UploadFile(params slack.UploadFileParameters) (*slack.FileSummary, error) {
	smc.ContentType = "file"
	smc.UploadFileCalls = append(smc.UploadFileCalls, params)
	return nil, nil
}

func (smc *SlackMockClient) PostMessage(channel string, opts ...slack.MsgOption) (string, string, error) {
	smc.ContentType = "message"
	smc.PostMessageCalls = append(smc.PostMessageCalls, channel)
	return channel, mockPostMessageTimestamp, nil
}

// chat.delete は削除したメッセージの channel と ts を返すため、モックは引数をそのまま返す
func (smc *SlackMockClient) DeleteMessage(channel, messageTimestamp string) (string, string, error) {
	smc.ContentType = "delete"
	smc.DeleteMessageCalls = append(smc.DeleteMessageCalls, MessageRef{channel: channel, timestamp: messageTimestamp})
	if smc.DeleteMessageErr != nil {
		return "", "", smc.DeleteMessageErr
	}
	return channel, messageTimestamp, nil
}
