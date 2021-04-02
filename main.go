package main

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	flag "github.com/spf13/pflag"
)

func main() {
	var (
		token     = os.Getenv("SLACK_TOKEN")
		channelID = flag.String("channel", "", "post slack channel id")
		text      = flag.String("text", "", "post text")
		iconEmoji = flag.String("icon", "", "icon emoji")
		userName  = flag.String("username", "", "user name")
	)
	flag.Parse()

	if token == "" {
		log.Fatal("error: SLACK_TOKEN is required")
	}
	if *channelID == "" {
		log.Fatal("error: --channel option is required")
	}
	if *text == "" {
		log.Fatal("error: --text option is required")
	}

	var (
		api  = slack.New(token)
		opts = []slack.MsgOption{
			slack.MsgOptionIconEmoji(*iconEmoji),
			slack.MsgOptionText(*text, false),
			slack.MsgOptionUsername(*userName),
		}
	)

	log.Println(
		api.PostMessage(
			*channelID,
			opts...,
		),
	)
}
