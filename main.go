package main

import (
	"fmt"
	"io"
	"io/ioutil"
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

type Options struct {
	token    string
	text     string
	filepath string

	mode        string
	snippetMode bool

	postOpts *PostOptions
}

type PostOptions struct {
	username  string
	channel   string
	iconEmoji string
	iconUrl   string
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

		noFail = flag.Bool("nofail", false, "always return success code(0)")

		errText []string
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version())
		os.Exit(0)
	}

	opts := &Options{
		snippetMode: *snippetMode,
		filepath:    *filepath,
		postOpts: &PostOptions{
			username:  *username,
			channel:   *channelID,
			iconEmoji: *iconEmoji,
			iconUrl:   *iconUrl,
		},
	}

	// token
	switch {
	case *optToken != "":
		opts.token = *optToken
	case envToken != "":
		opts.token = envToken
	default:
		errText = append(errText, "error: SLACK_TOKEN env or --token option is required")
	}

	if *channelID == "" {
		errText = append(errText, "error: --channel option is required")
	}

	// post mode
	switch {
	case *optText != "":
		opts.text = strings.Replace(*optText, "\\n", "\n", -1)
		opts.mode = "text"
	case *optTextFile != "":
		bytes, err := ioutil.ReadFile(*optTextFile)
		if err != nil {
			errText = append(errText, fmt.Sprintf("error: failed read text file: %s", err))
		}
		opts.text = string(bytes)
		opts.mode = "text"
	case *filepath != "":
		opts.filepath = *filepath
		opts.mode = "file"
	default:
		errText = append(errText, "error: --text option is required")
	}

	if 0 < len(errText) {
		fmt.Println(strings.Join(errText, "\n"))
		if *noFail {
			os.Exit(0)
		}
		os.Exit(1)
	}

	err := Do(opts)
	if err != nil {
		fmt.Println(err.Error())
		if *noFail {
			os.Exit(0)
		}
		os.Exit(1)
	}
	return
}

func Do(opts *Options) error {
	slackClient := slack.New(opts.token)

	switch opts.mode {
	case "text":
		// 3000文字以上は自動でスニペットにする
		// https://api.slack.com/reference/block-kit/blocks#section_fields
		if 3000 < len(opts.text) {
			opts.snippetMode = true
			fmt.Println("[INFO] text length exceed limits (3000 characters), upload snippets.")
		}

		if opts.snippetMode {
			if err := postFile(slackClient, opts.postOpts, strings.NewReader(opts.text), ""); err != nil {
				return errors.Wrap(err, "error postFile")
			}
		} else {
			if err := postMessage(slackClient, opts.postOpts, opts.text); err != nil {
				return errors.Wrap(err, "error: postMessage")
			}
		}
	case "file":
		if opts.filepath != "" {
			f, err := os.Open(opts.filepath)
			if err != nil {
				return errors.Wrapf(err, "error open file: %s", opts.filepath)
			}
			if err := postFile(slackClient, opts.postOpts, f, ""); err != nil {
				return errors.Wrapf(err, "error postFile %s", opts.filepath)
			}
		}
	}

	return nil
}

func (p *PostOptions) getMsgOptions() []slack.MsgOption {
	var opts []slack.MsgOption

	switch {
	case p.iconEmoji != "":
		opts = append(opts, slack.MsgOptionIconEmoji(p.iconEmoji))
	case p.iconUrl != "":
		opts = append(opts, slack.MsgOptionIconURL(p.iconUrl))
	}

	if p.username != "" {
		opts = append(opts, slack.MsgOptionUsername(p.username))
	}
	return opts
}

func postMessage(client *slack.Client, postOpts *PostOptions, text string) error {
	opts := []slack.MsgOption{}
	opts = append(opts, postOpts.getMsgOptions()...)
	opts = append(opts, slack.MsgOptionText(text, false))

	// TODO: timestampはThread等に使いまわしができるので、いつか出力したい
	_, ts, err := client.PostMessage(
		postOpts.channel,
		opts...,
	)
	_ = ts
	if err != nil {
		return errors.Wrap(err, "error postMessage")
	}
	return nil
}

func postFile(client *slack.Client, postOpts *PostOptions, fileReader io.Reader, comment string) error {
	fups := slack.FileUploadParameters{
		Filename:       "slack-quickpost",
		Reader:         fileReader,
		Filetype:       "auto",
		InitialComment: comment,
		Channels:       []string{postOpts.channel},
	}
	if _, err := client.UploadFile(fups); err != nil {
		return errors.Wrap(err, "error postFile")
	}
	return nil
}
