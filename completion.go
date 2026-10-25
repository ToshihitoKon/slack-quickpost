package main

import (
	_ "embed"

	"github.com/pkg/errors"
)

var (
	//go:embed completions/slack-quickpost.bash
	bashCompletion string
	//go:embed completions/slack-quickpost.zsh
	zshCompletion string
)

var completionScripts = map[string]string{
	"bash": bashCompletion,
	"zsh":  zshCompletion,
}

// 補完スクリプトは stdout を eval やファイル保存でシェルに読ませて使うため、未対応のシェル名を JSON の error として
// stdout に出さず、Set で弾いて stderr に出るフラグのパースエラーにする
type completionShellValue struct {
	shell string
}

func (v *completionShellValue) Set(s string) error {
	if _, ok := completionScripts[s]; !ok {
		return errors.New("unsupported shell (want bash or zsh)")
	}
	v.shell = s
	return nil
}

func (v *completionShellValue) String() string { return v.shell }
func (v *completionShellValue) Type() string   { return "string" }
