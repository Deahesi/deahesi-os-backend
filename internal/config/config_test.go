package config

import "testing"

func TestOpenRouterAPIKeyCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, standard, legacy, want string
	}{
		{name: "legacy", legacy: "legacy-key", want: "legacy-key"},
		{name: "standard", standard: "standard-key", want: "standard-key"},
		{name: "standard takes precedence", standard: "standard-key", legacy: "legacy-key", want: "standard-key"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OPENROUTER_API_KEY", tc.standard)
			t.Setenv("OPEN_ROUTER_API_KEY", tc.legacy)
			if got := NewConfig().LLM.OpenRouterAPI; got != tc.want {
				t.Errorf("got API key %q, want %q", got, tc.want)
			}
		})
	}
}
