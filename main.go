package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	flag "github.com/spf13/pflag"
)

// ビルド時に -ldflags "-X main.embeddedToken=... -X main.embeddedChannel=..." で
// 値を焼き込むための変数。通常のビルドでは空文字のまま。
// cmd/slack-quickpost-embed がこの変数へ値を注入する。
var (
	embeddedToken   string
	embeddedChannel string
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
	stdout      io.Writer

	token    string
	text     string
	filepath string
	blocks   string

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

func strGetFirstOne(vars ...string) string {
	for _, v := range vars {
		if v != "" {
			return v
		}
	}
	return ""
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWith(args, stdout, stderr, func(token string) SlackClient {
		return slack.New(token)
	})
}

func runWith(args []string, stdout, stderr io.Writer, newClient func(token string) SlackClient) int {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		// mode: print version
		printVersion = fs.Bool("version", false, "print version")

		// mode: post text
		optText     = fs.String("text", "", "post text")
		optTextFile = fs.String("textfile", "", "post text file path")
		snippetMode = fs.Bool("snippet", false, "post text as snippet")

		// mode: post file
		filepath = fs.String("file", "", "post file path")

		// mode: post blocks json
		optBlocks = fs.String("blocks", "", "post BlockKit json")

		// must options
		envToken   = os.Getenv("SLACK_TOKEN")
		optToken   = fs.String("token", "", "slack app OAuth token")
		optChannel = fs.String("channel", "", "post slack channel id")

		// optional
		envProfile = os.Getenv("SLACK_QUICKPOST_PROFILE")
		optProfile = fs.String("profile", "", "slack quickpost profile name")
		threadTs   = fs.String("thread-ts", "", "post under thread")
		iconEmoji  = fs.String("icon", "", "icon emoji")
		iconUrl    = fs.String("icon-url", "", "icon image url")
		username   = fs.String("username", "", "user name")

		noFail = fs.Bool("nofail", false, "always return success code(0)")

		errText []string
	)
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		fmt.Fprintln(stderr, err)
		return 2
	}

	if *printVersion {
		fmt.Fprintln(stdout, version())
		return 0
	}

	opts := &Options{
		stdout:      stdout,
		snippetMode: *snippetMode,
		filepath:    *filepath,
		postOpts: &PostOptions{
			username:  *username,
			iconEmoji: *iconEmoji,
			iconUrl:   *iconUrl,
			threadTs:  *threadTs,
		},
	}
	usr, err := user.Current()
	if err != nil {
		log.Printf("error: user.Current(). %s", err)
		return 1
	}

	profileName := strGetFirstOne(*optProfile, envProfile)

	var profile = &Profile{}
	if profileName != "" {
		profPath := path.Join(usr.HomeDir, ".config", "slack-quickpost", profileName+".yaml")
		profile, err = parseProfile(profPath)
		if err != nil {
			errText = append(errText, fmt.Sprintf("error: failed read profile %s. %s", profPath, err.Error()))
		}
	}

	opts.token = strGetFirstOne(*optToken, envToken, profile.Token, embeddedToken)
	if opts.token == "" {
		errText = append(errText, "error: slack token is required")
	}

	opts.postOpts.channel = strGetFirstOne(*optChannel, profile.Channel, embeddedChannel)
	if opts.postOpts.channel == "" {
		errText = append(errText, "error: channel is required")
	}

	// post mode
	switch {
	case *optText != "":
		opts.text = strings.Replace(*optText, "\\n", "\n", -1)
		opts.mode = "text"
	case *optBlocks != "":
		opts.blocks = *optBlocks
		opts.mode = "blocks"
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
		fmt.Fprintln(stdout, strings.Join(errText, "\n"))
		if *noFail {
			return 0
		}
		return 1
	}

	opts.slackClient = newClient(opts.token)

	output, err := Do(opts)
	if err != nil {
		fmt.Fprintln(stdout, err.Error())
		if *noFail {
			return 0
		}
		return 1
	}

	b, err := json.Marshal(output)
	if err != nil {
		fmt.Fprintln(stdout, err.Error())
		if *noFail {
			return 0
		}
		return 1
	}

	fmt.Fprintf(stdout, "%s", b)
	return 0
}

func Do(opts *Options) (*CliOutput, error) {
	var err error
	var output *CliOutput

	w := opts.stdout
	if w == nil {
		w = os.Stdout
	}

	switch opts.mode {
	case "text":
		// 3000文字以上は自動でスニペットにする
		// https://api.slack.com/reference/block-kit/blocks#section_fields
		if 3000 < len(opts.text) {
			opts.snippetMode = true
			fmt.Fprintln(w, "[INFO] text length exceed limits (3000 characters), upload snippets.")
		}

		if opts.snippetMode {
			output, err = postFile(opts.slackClient, opts.postOpts, strings.NewReader(opts.text), "", "", len(opts.text))
			if err != nil {
				return nil, errors.Wrap(err, "error postFile")
			}
		} else {
			output, err = postMessage(opts.slackClient, opts.postOpts, opts.text)
			if err != nil {
				return nil, errors.Wrap(err, "error: postMessage")
			}
		}
	case "blocks":
		blocks := slack.Blocks{}
		if err := blocks.UnmarshalJSON([]byte(opts.blocks)); err != nil {
			return nil, errors.Wrap(err, "error: failed blocks.UnmarshalJSON")
		}
		output, err = postBlocks(opts.slackClient, opts.postOpts, blocks)
		if err != nil {
			return nil, errors.Wrapf(err, "error postBlocks %s", opts.filepath)
		}

	case "file":
		if opts.filepath != "" {
			file, err := os.Open(opts.filepath)
			if err != nil {
				return nil, errors.Wrapf(err, "error open file: %s", opts.filepath)
			}
			filename := filepath.Base(opts.filepath)
			st, err := os.Stat(opts.filepath)
			if err != nil {
				return nil, errors.Wrapf(err, "error stat file: %s", opts.filepath)
			}
			output, err = postFile(opts.slackClient, opts.postOpts, file, filename, "", int(st.Size()))
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

func postBlocks(client SlackClient, postOpts *PostOptions, blocks slack.Blocks) (*CliOutput, error) {
	opts := []slack.MsgOption{}
	opts = append(opts, postOpts.getMsgOptions()...)
	opts = append(opts, slack.MsgOptionBlocks(blocks.BlockSet...))

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

func postFile(client SlackClient, postOpts *PostOptions, fileReader io.Reader, filename, comment string, size int) (*CliOutput, error) {
	postTime := time.Now()
	if filename == "" {
		filename = fmt.Sprintf("%s.txt", postTime.Format("20060102_150405.999999"))
	}
	fups := slack.UploadFileParameters{
		Filename:        filename,
		FileSize:        size,
		Reader:          fileReader,
		InitialComment:  comment,
		Channel:         postOpts.channel,
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
