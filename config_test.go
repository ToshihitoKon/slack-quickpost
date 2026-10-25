package main

import "testing"

func Test_profilePath_acceptsNameWithOrWithoutExtension(t *testing.T) {
	tests := []struct {
		title string
		name  string
		want  string
	}{
		{"拡張子なし", "work", "/home/u/.config/slack-quickpost/work.yaml"},
		{"拡張子付き", "work.yaml", "/home/u/.config/slack-quickpost/work.yaml"},
		{".yml は外さない", "work.yml", "/home/u/.config/slack-quickpost/work.yml.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if got := profilePath("/home/u", tt.name); got != tt.want {
				t.Errorf("profilePath(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
