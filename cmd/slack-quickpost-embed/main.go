// Command slack-quickpost-embed builds a slack-quickpost binary with a Slack
// token and default channel baked in via -ldflags.
//
// 焼き込んだバイナリは、実行時に --token / --channel / SLACK_TOKEN / プロファイルの
// いずれも指定されなかったときのフォールバックとして焼き込み値を使う。
// つまり実行時の明示的な指定が常に優先される。
package main

import (
	"fmt"
	"os"
	"os/exec"
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

	target, err := buildTarget(*optSource, *optVersion)
	if err != nil {
		return err
	}

	ldflags := fmt.Sprintf("-X %s=%s -X %s=%s", tokenVar, token, channelVar, channel)
	args := []string{"build", "-o", *optOutput, "-ldflags", ldflags, target}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	fmt.Printf("built %s (channel=%s, token embedded)\n", *optOutput, channel)
	return nil
}

// buildTarget は go build に渡すビルド対象を組み立てる。
// module: バージョン付きモジュールパス、local: カレントのパッケージ。
func buildTarget(source, version string) (string, error) {
	switch source {
	case "module":
		return fmt.Sprintf("%s@%s", modulePath, version), nil
	case "local":
		return ".", nil
	default:
		return "", fmt.Errorf("invalid --source %q (want \"module\" or \"local\")", source)
	}
}

func strGetFirstOne(vars ...string) string {
	for _, v := range vars {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
