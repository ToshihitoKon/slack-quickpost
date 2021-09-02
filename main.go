package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/slack-go/slack"
	flag "github.com/spf13/pflag"
)

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Goモジュールが無効など
		return "(devel)"
	}
	return info.Main.Version
}

func main() {
	var (
		token        = flag.String("token", "", "slack app OAuth token")
		channelID    = flag.String("channel", "", "post slack channel id")
		text         = flag.String("text", "", "post text")
		iconEmoji    = flag.String("icon", "", "icon emoji")
		iconUrl      = flag.String("icon-url", "", "icon image url")
		userName     = flag.String("username", "", "user name")
		printVersion = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version())
		os.Exit(0)
	}

	var errText []string
	if *token == "" {
		*token = os.Getenv("SLACK_TOKEN")
		if *token == "" {
			errText = append(errText, "error: SLACK_TOKEN env or --token option is required")
		}
	}
	if *channelID == "" {
		errText = append(errText, "error: --channel option is required")
	}
	if *text == "" {
		errText = append(errText, "error: --text option is required")
	}
	if 0 < len(errText) {
		fmt.Println(strings.Join(errText, "\n"))
		os.Exit(1)
	}

	var (
		api  = slack.New(*token)
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

	channel, ts, err := api.PostMessage(
		*channelID,
		opts...,
	)
	if err != nil {
		log.Fatal("error: slack.PostMessage failed", err)
	}
	fmt.Println("success", channel, ts)
}
