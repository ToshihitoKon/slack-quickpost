package main

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Token   string `yaml:"token"`
	Channel string `yaml:"channel"`
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
