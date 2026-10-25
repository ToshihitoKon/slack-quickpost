package main

import (
	"io"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Token   string `yaml:"token"`
	Channel string `yaml:"channel"`
}

// .yml は外さない。プロファイルは .yaml しか読まないため、foo.yml を foo.yaml に読み替えると別のファイルを読むことになる
func profilePath(homeDir, name string) string {
	return path.Join(homeDir, ".config", "slack-quickpost", strings.TrimSuffix(name, ".yaml")+".yaml")
}

func parseProfile(filepath string) (*Profile, error) {
	prf := &Profile{}
	f, err := os.Open(filepath)

	if err != nil {
		return prf, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return prf, err
	}

	if err := yaml.Unmarshal(data, prf); err != nil {
		return prf, err
	}

	return prf, nil
}
