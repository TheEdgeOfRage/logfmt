package config

import "testing"

func TestSetFilterWithEquals(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setFilter("token=abc=123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Filter["token"]; got != "abc=123" {
		t.Fatalf("got %q, want %q", got, "abc=123")
	}
}
