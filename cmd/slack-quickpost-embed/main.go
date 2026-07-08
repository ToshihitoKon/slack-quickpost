// Command slack-quickpost-embed builds a slack-quickpost binary with a Slack
// token and default channel baked in via -ldflags.
//
// 焼き込んだバイナリは、実行時に --token / --channel / SLACK_TOKEN / プロファイルの
// いずれも指定されなかったときのフォールバックとして焼き込み値を使う。
// つまり実行時の明示的な指定が常に優先される。
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

// slack-quickpost 本体のモジュールパス。module ソースをビルドする際に使う。
const modulePath = "github.com/ToshihitoKon/slack-quickpost"

// 本体側で焼き込み対象となっている変数のシンボル名。
const (
	tokenVar   = "main.embeddedToken"
	channelVar = "main.embeddedChannel"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		optToken   = flag.String("token", "", "slack app OAuth token to embed (fallback: SLACK_TOKEN)")
		optChannel = flag.String("channel", "", "default slack channel id to embed (fallback: SLACK_CHANNEL)")
		optOutput  = flag.String("output", "slack-quickpost", "output binary path")
		optSource  = flag.String("source", "module", "build source: \"module\" (fetch by version) or \"local\" (current directory)")
		optVersion = flag.String("version", "latest", "slack-quickpost version to build when --source=module")
	)
	flag.Parse()

	token := strGetFirstOne(*optToken, os.Getenv("SLACK_TOKEN"))
	channel := strGetFirstOne(*optChannel, os.Getenv("SLACK_CHANNEL"))

	if token == "" {
		return fmt.Errorf("token is required (--token or SLACK_TOKEN)")
	}
	if channel == "" {
		return fmt.Errorf("channel is required (--channel or SLACK_CHANNEL)")
	}

	ldflags := fmt.Sprintf("-X %s=%s -X %s=%s", tokenVar, token, channelVar, channel)

	var err error
	switch *optSource {
	case "module":
		err = buildModule(*optVersion, ldflags, *optOutput)
	case "local":
		err = buildLocal(ldflags, *optOutput)
	default:
		return fmt.Errorf("invalid --source %q (want \"module\" or \"local\")", *optSource)
	}
	if err != nil {
		return err
	}

	fmt.Printf("built %s (channel=%s, token embedded)\n", *optOutput, channel)
	return nil
}

// buildModule は公開モジュールを取得して焼き込みビルドする。
// path@version 構文は go install でしか使えず、かつ go install は -o で出力先を
// 指定できないため、一時ディレクトリを GOBIN に指定してインストールし、
// 生成物を output へ移す。go.mod のない場所からでも実行できる。
func buildModule(version, ldflags, output string) error {
	gobin, err := os.MkdirTemp("", "slack-quickpost-embed-")
	if err != nil {
		return fmt.Errorf("create temp GOBIN: %w", err)
	}
	defer os.RemoveAll(gobin)

	target := fmt.Sprintf("%s@%s", modulePath, version)
	cmd := exec.Command("go", "install", "-ldflags", ldflags, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// GOBIN を一時ディレクトリに向けて、ユーザーの $GOPATH/bin を汚さない。
	cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	// go install はモジュール名末尾をバイナリ名にする。
	built := filepath.Join(gobin, filepath.Base(modulePath))
	if err := moveFile(built, output); err != nil {
		return fmt.Errorf("move binary to %s: %w", output, err)
	}
	return nil
}

// buildLocal はカレントの slack-quickpost ソースを焼き込みビルドする。
// go.mod のあるチェックアウト内で実行される前提。
func buildLocal(ldflags, output string) error {
	cmd := exec.Command("go", "build", "-o", output, "-ldflags", ldflags, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

// moveFile は src を dst へ移動する。GOBIN の一時ディレクトリと出力先が別ファイル
// システムだと os.Rename が EXDEV で失敗しうるため、その場合はコピーにフォールバックする。
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func strGetFirstOne(vars ...string) string {
	for _, v := range vars {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
