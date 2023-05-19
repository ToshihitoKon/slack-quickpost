package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

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

type CliOutput struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"`
}

type Options struct {
	slackClient SlackClient

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
	threadTs  string
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
		threadTs  = flag.String("thread-ts", "", "post under thread")
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
			threadTs:  *threadTs,
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

	opts.slackClient = slack.New(opts.token)

	output, err := Do(opts)
	if err != nil {
		fmt.Println(err.Error())
		if *noFail {
			os.Exit(0)
		}
		os.Exit(1)
	}

	b, err := json.Marshal(output)
	if err != nil {
		fmt.Println(err.Error())
		if *noFail {
			os.Exit(0)
		}
		os.Exit(1)
	}

	fmt.Printf("%s", b)
	return
}

func Do(opts *Options) (*CliOutput, error) {
	var err error
	var output *CliOutput

	switch opts.mode {
	case "text":
		// 3000文字以上は自動でスニペットにする
		// https://api.slack.com/reference/block-kit/blocks#section_fields
		if 3000 < len(opts.text) {
			opts.snippetMode = true
			fmt.Println("[INFO] text length exceed limits (3000 characters), upload snippets.")
		}

		if opts.snippetMode {
			output, err = postFile(opts.slackClient, opts.postOpts, strings.NewReader(opts.text), "", "")
			if err != nil {
				return nil, errors.Wrap(err, "error postFile")
			}
		} else {
			output, err = postMessage(opts.slackClient, opts.postOpts, opts.text)
			if err != nil {
				return nil, errors.Wrap(err, "error: postMessage")
			}
		}
	case "file":
		if opts.filepath != "" {
			file, err := os.Open(opts.filepath)
			if err != nil {
				return nil, errors.Wrapf(err, "error open file: %s", opts.filepath)
			}
			filename := filepath.Base(opts.filepath)
			output, err = postFile(opts.slackClient, opts.postOpts, file, filename, "")
			if err != nil {
				return nil, errors.Wrapf(err, "error postFile %s", opts.filepath)
			}
		}
	}

	return output, nil
}

func (p *PostOptions) getMsgOptions() []slack.MsgOption {
	var opts []slack.MsgOption

	if p.threadTs != "" {
		opts = append(opts, slack.MsgOptionTS(p.threadTs))
	}

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

func postMessage(client SlackClient, postOpts *PostOptions, text string) (*CliOutput, error) {
	opts := []slack.MsgOption{}
	opts = append(opts, postOpts.getMsgOptions()...)
	opts = append(opts, slack.MsgOptionText(text, false))

	postedChannel, ts, err := client.PostMessage(
		postOpts.channel,
		opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error postMessage")
	}

	output := &CliOutput{
		Channel:   postedChannel,
		Timestamp: ts,
	}

	return output, nil
}

func postFile(client SlackClient, postOpts *PostOptions, fileReader io.Reader, filename, comment string) (*CliOutput, error) {
	if filename == "" {
		filename = fmt.Sprintf("%s.txt", time.Now().Format("20060102_150405"))
	}
	fups := slack.FileUploadParameters{
		Filename:        filename,
		Reader:          fileReader,
		Filetype:        "auto",
		InitialComment:  comment,
		Channels:        []string{postOpts.channel},
		ThreadTimestamp: postOpts.threadTs,
	}
	if _, err := client.UploadFile(fups); err != nil {
		return nil, errors.Wrap(err, "error postFile")
	}
	output := &CliOutput{
		Channel: postOpts.channel,
		// Timestamp: UploadFileはtimestampを取得できない
	}
	return output, nil
}
