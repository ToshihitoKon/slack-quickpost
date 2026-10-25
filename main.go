package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

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
	Channel   string `json:"channel,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Error     string `json:"error,omitempty"`
}

const (
	modeText   = "text"
	modeBlocks = "blocks"
	modeFile   = "file"
	modeDelete = "delete"
)

type Options struct {
	slackClient SlackClient
	infoOut     io.Writer

	token    string
	text     string
	filepath string
	blocks   string

	mode        string
	snippetMode bool

	postOpts     *PostOptions
	deleteTarget MessageRef
}

type PostOptions struct {
	username  string
	channel   string
	iconEmoji string
	iconUrl   string
	threadTs  string
}

type MessageRef struct {
	channel   string
	timestamp string
}

// permalink の p 以降はメッセージの ts から "." を除いた数字列で、末尾 6 桁が小数部にあたる。
var permalinkPathPattern = regexp.MustCompile(`^/archives/([A-Z0-9]+)/p(\d+)(\d{6})$`)

const permalinkFormat = "https://<workspace>.slack.com/archives/<channel>/p<timestamp>"

// parseMessageURL は Slack の「リンクをコピー」で得られる permalink から channel と ts を取り出す。
// スレッド返信の permalink に付く thread_ts クエリは親の ts なので使わない。
// 削除対象は channel と ts だけで決まるため、ホスト部は検証しない。
func parseMessageURL(rawURL string) (MessageRef, error) {
	var m []string
	if u, err := url.Parse(rawURL); err == nil {
		m = permalinkPathPattern.FindStringSubmatch(u.Path)
	}
	if m == nil {
		return MessageRef{}, fmt.Errorf("invalid message url (want a permalink like %s): %s", permalinkFormat, rawURL)
	}
	return MessageRef{
		channel:   m[1],
		timestamp: m[2] + "." + m[3],
	}, nil
}

// --delete はこれら以外のフラグと排他。後から追加されるフラグも既定で排他になるよう許可リストで持つ。
var deleteCompatibleFlags = map[string]bool{
	"delete":  true,
	"token":   true,
	"profile": true,
	"nofail":  true,
}

// pflag の String は重複指定を黙って後勝ちにし、StringArray は --help の型表示が stringArray になって
// 複数指定できるように見えるため、自前で指定回数を数える
type countingStringValue struct {
	value string
	count int
}

func (v *countingStringValue) Set(s string) error {
	v.value = s
	v.count++
	return nil
}

func (v *countingStringValue) String() string { return v.value }
func (v *countingStringValue) Type() string   { return "string" }

func flagsConflictingWithDelete(fs *flag.FlagSet) []string {
	var names []string
	fs.Visit(func(f *flag.Flag) {
		if !deleteCompatibleFlags[f.Name] {
			names = append(names, "--"+f.Name)
		}
	})
	return names
}

// 投稿モードを選ぶフラグ。複数指定を優先順で片方だけ採用すると指定ミスに気づけないため、排他にする
var postModeFlags = []string{"text", "textfile", "blocks", "file"}

func specifiedFlags(fs *flag.FlagSet, names []string) []string {
	var specified []string
	for _, name := range names {
		if fs.Changed(name) {
			specified = append(specified, "--"+name)
		}
	}
	return specified
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

// writeError は失敗時の出力契約（stdoutにJSON1行、--nofailはexit codeのみ変える）を一箇所に集約する。
// Encodeの書き込みエラーはここでは無視する。既にエラー経路にいるため報告先がなく、
// --nofailの有無に関わらずexit codeは変わらない。
func writeError(stdout io.Writer, msg string, noFail bool) int {
	_ = json.NewEncoder(stdout).Encode(&CliOutput{Error: msg})
	if noFail {
		return 0
	}
	return 1
}

func runWith(args []string, stdout, stderr io.Writer, newClient func(token string) SlackClient) int {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		// mode: print version
		printVersion = fs.Bool("version", false, "print version")

		// mode: print shell completion script
		optCompletion = &completionShellValue{}

		// mode: post text
		optText     = fs.String("text", "", "post text")
		optTextFile = fs.String("textfile", "", "post text file path")
		snippetMode = fs.Bool("snippet", false, "post text as snippet")

		// mode: post file
		optFile = fs.String("file", "", "post file path")

		// mode: post blocks json
		optBlocks = fs.String("blocks", "", "post BlockKit json")

		// mode: delete message
		optDelete = &countingStringValue{}

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

		noFail = fs.Bool("nofail", false, "exit with code 0 even on failure (except flag parse errors)")

		errText []string
	)
	fs.Var(optDelete, "delete", "delete message by permalink url")
	fs.Var(optCompletion, "completion", "print shell completion script (bash or zsh)")

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

	if optCompletion.shell != "" {
		fmt.Fprint(stdout, completionScripts[optCompletion.shell])
		return 0
	}

	// 位置引数は使わないが、黙って捨てると、値を取るフラグが次のフラグを値として食った書き間違い（--text --delete URL）に気づけない
	if 0 < fs.NArg() {
		errText = append(errText, fmt.Sprintf("unexpected arguments: %s", strings.Join(fs.Args(), " ")))
	}

	opts := &Options{
		infoOut:     stderr,
		snippetMode: *snippetMode,
		postOpts: &PostOptions{
			username:  *username,
			iconEmoji: *iconEmoji,
			iconUrl:   *iconUrl,
			threadTs:  *threadTs,
		},
	}
	profile := &Profile{}
	if profileName := strGetFirstOne(*optProfile, envProfile); profileName != "" {
		loaded, err := loadProfile(profileName)
		if err != nil {
			errText = append(errText, err.Error())
		} else {
			profile = loaded
		}
	}

	opts.token = strGetFirstOne(*optToken, envToken, profile.Token, embeddedToken)
	if opts.token == "" {
		errText = append(errText, "slack token is required")
	}

	// delete mode は channel を URL から取るため、profile や焼き込みの channel は使わない
	if 0 < optDelete.count {
		opts.mode = modeDelete
		if conflicting := flagsConflictingWithDelete(fs); 0 < len(conflicting) {
			errText = append(errText, fmt.Sprintf("--delete cannot be used with %s", strings.Join(conflicting, ", ")))
		}
		if 1 < optDelete.count {
			errText = append(errText, "--delete can be specified only once")
		}
		target, err := parseMessageURL(optDelete.value)
		if err != nil {
			errText = append(errText, err.Error())
		}
		opts.deleteTarget = target
	} else {
		opts.postOpts.channel = strGetFirstOne(*optChannel, profile.Channel, embeddedChannel)
		if opts.postOpts.channel == "" {
			errText = append(errText, "channel is required")
		}

		if modes := specifiedFlags(fs, postModeFlags); 1 < len(modes) {
			errText = append(errText, fmt.Sprintf("%s cannot be used together", strings.Join(modes, ", ")))
		}

		// post mode
		switch {
		case *optText != "":
			opts.text = strings.ReplaceAll(*optText, "\\n", "\n")
			opts.mode = modeText
		case *optBlocks != "":
			opts.blocks = *optBlocks
			opts.mode = modeBlocks
		case *optTextFile != "":
			bytes, err := os.ReadFile(*optTextFile)
			if err != nil {
				errText = append(errText, fmt.Sprintf("failed to read text file: %s", err))
			}
			opts.text = string(bytes)
			opts.mode = modeText
		case *optFile != "":
			opts.filepath = *optFile
			opts.mode = modeFile
		default:
			errText = append(errText, "one of --text, --textfile, --blocks or --file is required")
		}
	}

	if 0 < len(errText) {
		return writeError(stdout, strings.Join(errText, "\n"), *noFail)
	}

	opts.slackClient = newClient(opts.token)

	output, err := Do(opts)
	if err != nil {
		return writeError(stdout, err.Error(), *noFail)
	}

	// 投稿・削除は完了しているため、書き込み失敗で失敗のexit codeを返すと、リトライ時に二重投稿や削除済みメッセージへの削除エラーになる
	if err := json.NewEncoder(stdout).Encode(output); err != nil {
		fmt.Fprintf(stderr, "failed to write output: %s\n", err)
	}
	return 0
}

func Do(opts *Options) (*CliOutput, error) {
	switch opts.mode {
	case modeText:
		// 3000バイトを超えると自動でスニペットにする（Slackの上限は3000文字だが、判定はバイト数で行う）
		// https://api.slack.com/reference/block-kit/blocks#section_fields
		isOverLimit := 3000 < len(opts.text)
		if isOverLimit {
			fmt.Fprintln(opts.infoOut, "[INFO] text length exceed limits (3000 bytes), upload snippets.")
		}

		if opts.snippetMode || isOverLimit {
			output, err := postFile(opts.slackClient, opts.postOpts, strings.NewReader(opts.text), "", len(opts.text))
			if err != nil {
				return nil, fmt.Errorf("failed to upload file: %w", err)
			}
			return output, nil
		}
		output, err := postMessage(opts.slackClient, opts.postOpts, slack.MsgOptionText(opts.text, false))
		if err != nil {
			return nil, fmt.Errorf("failed to post message: %w", err)
		}
		return output, nil

	case modeBlocks:
		blocks := slack.Blocks{}
		if err := blocks.UnmarshalJSON([]byte(opts.blocks)); err != nil {
			return nil, fmt.Errorf("failed to parse blocks json: %w", err)
		}
		output, err := postMessage(opts.slackClient, opts.postOpts, slack.MsgOptionBlocks(blocks.BlockSet...))
		if err != nil {
			return nil, fmt.Errorf("failed to post blocks: %w", err)
		}
		return output, nil

	case modeFile:
		file, err := os.Open(opts.filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %s: %w", opts.filepath, err)
		}
		defer file.Close()
		st, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %s: %w", opts.filepath, err)
		}
		output, err := postFile(opts.slackClient, opts.postOpts, file, filepath.Base(opts.filepath), int(st.Size()))
		if err != nil {
			return nil, fmt.Errorf("failed to upload file: %s: %w", opts.filepath, err)
		}
		return output, nil

	case modeDelete:
		output, err := deleteMessage(opts.slackClient, opts.deleteTarget)
		if err != nil {
			return nil, fmt.Errorf("failed to delete message: %w", err)
		}
		return output, nil
	}

	return nil, fmt.Errorf("unknown mode: %q", opts.mode)
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

// content には本文（slack.MsgOptionText）か BlockKit（slack.MsgOptionBlocks）を渡す
func postMessage(client SlackClient, postOpts *PostOptions, content slack.MsgOption) (*CliOutput, error) {
	opts := append(postOpts.getMsgOptions(), content)

	postedChannel, ts, err := client.PostMessage(postOpts.channel, opts...)
	if err != nil {
		return nil, err
	}

	output := &CliOutput{
		Channel:   postedChannel,
		Timestamp: ts,
	}
	return output, nil
}

func postFile(client SlackClient, postOpts *PostOptions, fileReader io.Reader, filename string, size int) (*CliOutput, error) {
	if filename == "" {
		filename = fmt.Sprintf("%s.txt", time.Now().Format("20060102_150405.999999"))
	}
	fups := slack.UploadFileParameters{
		Filename:        filename,
		FileSize:        size,
		Reader:          fileReader,
		Channel:         postOpts.channel,
		ThreadTimestamp: postOpts.threadTs,
	}
	if _, err := client.UploadFile(fups); err != nil {
		return nil, err
	}
	output := &CliOutput{
		// UploadFile のレスポンス(FileSummary)にはchannel情報が含まれないため、
		// リクエスト時点のchannelをそのまま使う
		Channel: postOpts.channel,
		// Timestamp: UploadFileはtimestampを取得できない
	}
	return output, nil
}

func deleteMessage(client SlackClient, target MessageRef) (*CliOutput, error) {
	deletedChannel, ts, err := client.DeleteMessage(target.channel, target.timestamp)
	if err != nil {
		return nil, err
	}

	output := &CliOutput{
		Channel:   deletedChannel,
		Timestamp: ts,
	}
	return output, nil
}
