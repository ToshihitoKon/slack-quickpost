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
		iconUrl   = flag.String("icon-url", "", "icon image url")
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
			slack.MsgOptionText(*text, false),
			slack.MsgOptionUsername(*userName),
		}
	)

	if *iconEmoji != "" {
		opts = append(opts, slack.MsgOptionIconEmoji(*iconEmoji))
	}
	if *iconUrl != "" {
		opts = append(opts, slack.MsgOptionIconURL(*iconUrl))
	}

	log.Println(
		api.PostMessage(
			*channelID,
			opts...,
		),
	)
}
