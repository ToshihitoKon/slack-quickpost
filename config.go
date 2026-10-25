package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Token   string `yaml:"token"`
	Channel string `yaml:"channel"`
}

// .yml は外さない。プロファイルは .yaml しか読まないため、foo.yml を foo.yaml に読み替えると別のファイルを読むことになる
func profilePath(homeDir, name string) string {
	return filepath.Join(homeDir, ".config", "slack-quickpost", strings.TrimSuffix(name, ".yaml")+".yaml")
}

func loadProfile(name string) (*Profile, error) {
	// user.Current は cgo なしのビルドだと、UID が /etc/passwd になく $USER も未設定の環境で失敗するため、
	// $HOME を見る os.UserHomeDir を使う
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to read profile %s: %w", name, err)
	}
	path := profilePath(home, name)
	prf, err := parseProfile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile %s: %w", path, err)
	}
	return prf, nil
}

func parseProfile(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	prf := &Profile{}
	if err := yaml.Unmarshal(data, prf); err != nil {
		return nil, err
	}
	return prf, nil
}
