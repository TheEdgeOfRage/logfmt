package config

import "testing"

func TestParseFieldListTrimsSpaces(t *testing.T) {
	got := parseFieldList("msg, id , ,error")
	want := []string{"msg", "id", "error"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, field := range got {
		if field != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestSetFilterTrimsParts(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setFilter("msg = hello, id = 1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Filter["msg"]; got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
	if got := cfg.Filter["id"]; got != "1" {
		t.Fatalf("got %q, want %q", got, "1")
	}
}

func TestSetFilterWithEquals(t *testing.T) {
	cfg := &Config{}
	if err := cfg.setFilter("token=abc=123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Filter["token"]; got != "abc=123" {
		t.Fatalf("got %q, want %q", got, "abc=123")
	}
}
