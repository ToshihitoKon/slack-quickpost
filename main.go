package main

import (
	"flag"
	"log"
	"os"

	"github.com/slack-go/slack"
)

func main() {
	var (
		channelID = flag.String("c", "", "post slack channel id")
		text      = flag.String("t", "", "post text")
		iconEmoji = flag.String("i", "", "icon emoji")
		userName  = flag.String("u", "", "user name")
	)
	flag.Parse()

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatal("SLACK_TOKEN is required")
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
