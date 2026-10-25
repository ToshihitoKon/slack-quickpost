package main

import (
	"path/filepath"
	"testing"
)

func Test_profilePath_acceptsNameWithOrWithoutExtension(t *testing.T) {
	tests := []struct {
		title    string
		name     string
		wantFile string
	}{
		{"拡張子なし", "work", "work.yaml"},
		{"拡張子付き", "work.yaml", "work.yaml"},
		{".yml は外さない", "work.yml", "work.yml.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			want := filepath.Join("/home/u", ".config", "slack-quickpost", tt.wantFile)
			if got := profilePath("/home/u", tt.name); got != want {
				t.Errorf("profilePath(%q) = %q, want %q", tt.name, got, want)
			}
		})
	}
}
