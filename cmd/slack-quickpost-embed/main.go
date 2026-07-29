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
	"path/filepath"
	"runtime"
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
		optGoos    = flag.String("goos", "", "target GOOS for cross compilation (default: host GOOS)")
		optGoarch  = flag.String("goarch", "", "target GOARCH for cross compilation (default: host GOARCH)")
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

	goos := strGetFirstOne(*optGoos, os.Getenv("GOOS"), runtime.GOOS)
	goarch := strGetFirstOne(*optGoarch, os.Getenv("GOARCH"), runtime.GOARCH)

	// buildModule は使い捨てモジュールを cmd.Dir 配下に作る都合上、相対パスのまま
	// 渡すと tmpdir 起点で解決されて defer os.RemoveAll(tmpdir) に巻き込まれてしまう。
	// 必ず絶対パス化してから両ビルド関数に渡す。
	output, err := filepath.Abs(*optOutput)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	if goos == "windows" && !strings.HasSuffix(output, ".exe") {
		output += ".exe"
	}

	ldflags := fmt.Sprintf("-s -w -X %s=%s -X %s=%s", tokenVar, token, channelVar, channel)

	switch *optSource {
	case "module":
		err = buildModule(*optVersion, ldflags, goos, goarch, output)
	case "local":
		err = buildLocal(ldflags, goos, goarch, output)
	default:
		return fmt.Errorf("invalid --source %q (want \"module\" or \"local\")", *optSource)
	}
	if err != nil {
		return err
	}

	fmt.Printf("built %s (channel=%s, token embedded)\n", output, channel)
	return nil
}

// buildModule は公開モジュールを取得して焼き込みビルドする。
// go install は GOBIN が設定された状態でのクロスコンパイルを許可しない
// （"cannot install cross-compiled binaries when GOBIN is set"）ため、
// 一時ディレクトリを使い捨てモジュールのルート（cmd.Dir）として使い、
// そこで `go get <module>@<version>` により対象モジュールを取り込んでから
// `go build -o` でビルドする。go.mod のない場所からでも実行できる。
func buildModule(version, ldflags, goos, goarch, output string) error {
	tmpdir, err := os.MkdirTemp("", "slack-quickpost-embed-")
	if err != nil {
		return fmt.Errorf("create temp module dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	initCmd := exec.Command("go", "mod", "init", "slack-quickpost-embed-tmp")
	initCmd.Dir = tmpdir
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("go mod init failed: %w", err)
	}

	target := fmt.Sprintf("%s@%s", modulePath, version)
	getCmd := exec.Command("go", "get", target)
	getCmd.Dir = tmpdir
	getCmd.Stdout = os.Stdout
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		return fmt.Errorf("go get failed: %w", err)
	}

	buildCmd := exec.Command("go", "build", "-trimpath", "-o", output, "-ldflags", ldflags, modulePath)
	buildCmd.Dir = tmpdir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

// buildLocal はカレントの slack-quickpost ソースを焼き込みビルドする。
// go.mod のあるチェックアウト内で実行される前提。
func buildLocal(ldflags, goos, goarch, output string) error {
	cmd := exec.Command("go", "build", "-trimpath", "-o", output, "-ldflags", ldflags, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

func strGetFirstOne(vars ...string) string {
	for _, v := range vars {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
