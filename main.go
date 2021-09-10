package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
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
		// mode: print version
		printVersion = flag.Bool("version", false, "print version")
		// mode: post text
		optText     = flag.String("text", "", "post text")
		optTextFile = flag.String("textfile", "", "post text file path")
		snippetMode = flag.Bool("snippet", false, "post text as snippet")
		// mode: post file
		filepath = flag.String("file", "", "post file path")

		// must options
		envToken  = os.Getenv("SLACK_TOKEN")
		optToken  = flag.String("token", "", "slack app OAuth token")
		channelID = flag.String("channel", "", "post slack channel id")

		// optional
		iconEmoji = flag.String("icon", "", "icon emoji")
		iconUrl   = flag.String("icon-url", "", "icon image url")
		username  = flag.String("username", "", "user name")

		token    string
		postText string
		errText  []string
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version())
		os.Exit(0)
	}

	switch {
	case *optToken != "":
		token = *optToken
	case envToken != "":
		token = envToken
	default:
		errText = append(errText, "error: SLACK_TOKEN env or --token option is required")
	}
	if *channelID == "" {
		errText = append(errText, "error: --channel option is required")
	}

	postOpts := postOptions{
		Username:  *username,
		Channel:   *channelID,
		IconEmoji: *iconEmoji,
		IconUrl:   *iconUrl,
	}
	mode := ""

	switch {
	case *optText != "":
		postText = strings.Replace(*optText, "\\n", "\n", -1)
		mode = "text"
	case *optTextFile != "":
		bytes, err := ioutil.ReadFile(*optTextFile)
		if err != nil {
			errText = append(errText, fmt.Sprintf("error: failed read text file: %s", err))
		}
		postText = string(bytes)
		mode = "text"
	case *filepath != "":
		mode = "file"
	default:
		errText = append(errText, "error: --text option is required")
	}
	if 0 < len(errText) {
		fmt.Println(strings.Join(errText, "\n"))
		os.Exit(1)
	}

	slackClient := slack.New(token)

	switch mode {
	case "text":
		// 3000文字以上は自動でスニペットにする
		// https://api.slack.com/reference/block-kit/blocks#section_fields
		if 3000 < len(postText) {
			*snippetMode = true
			fmt.Println("[INFO] text length exceed limits (3000 characters), upload snippets.")
		}

		if *snippetMode {
			if err := postFile(slackClient, postOpts, strings.NewReader(postText), ""); err != nil {
				log.Fatal("error: postFile ", err)
			}
		} else {
			if err := postMessage(slackClient, postOpts, postText); err != nil {
				log.Fatal("error: postMessage: ", err)
			}
		}
	case "file":
		if *filepath != "" {
			f, err := os.Open(*filepath)
			if err != nil {
				log.Fatal("error: open file, ", *filepath)
			}
			if err := postFile(slackClient, postOpts, f, ""); err != nil {
				log.Fatal("error: postFile ", err)
			}
		}
	}
}

type postOptions struct {
	Username  string
	Channel   string
	IconEmoji string
	IconUrl   string
}

func (p *postOptions) getMsgOptions() []slack.MsgOption {
	var opts []slack.MsgOption

	switch {
	case p.IconEmoji != "":
		opts = append(opts, slack.MsgOptionIconEmoji(p.IconEmoji))
	case p.IconUrl != "":
		opts = append(opts, slack.MsgOptionIconURL(p.IconUrl))
	}

	if p.Username != "" {
		opts = append(opts, slack.MsgOptionUsername(p.Username))
	}
	return opts
}

func postMessage(client *slack.Client, postOpts postOptions, text string) error {
	opts := []slack.MsgOption{}
	opts = append(opts, postOpts.getMsgOptions()...)
	opts = append(opts, slack.MsgOptionText(text, false))

	// TODO: timestampはThread等に使いまわしができるので、いつか出力したい
	_, ts, err := client.PostMessage(
		postOpts.Channel,
		opts...,
	)
	_ = ts
	if err != nil {
		return errors.Wrap(err, "error postMessage")
	}
	return nil
}

func postFile(client *slack.Client, postOpts postOptions, fileReader io.Reader, comment string) error {
	fups := slack.FileUploadParameters{
		Filename:       "slack-quickpost",
		Reader:         fileReader,
		Filetype:       "auto",
		InitialComment: comment,
		Channels:       []string{postOpts.Channel},
	}
	if _, err := client.UploadFile(fups); err != nil {
		return errors.Wrap(err, "error postFile")
	}
	return nil
}
