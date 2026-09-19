package config

import (
	"github.com/BurntSushi/toml"
	"testing"
)

func TestFusionAuthBypassIsExplicit(t *testing.T) {
	for _, tc := range []struct {
		text     string
		disabled bool
	}{
		{"[auth]\n", false},
		{"[auth]\ndisable_fusionauth = false\n", false},
		{"[auth]\ndisable_fusionauth = true\n", true},
	} {
		var cfg Config
		if err := toml.Unmarshal([]byte(tc.text), &cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Auth.DisableFusionAuth != tc.disabled {
			t.Fatal("incorrect bypass default or TOML mapping")
		}
	}
}
