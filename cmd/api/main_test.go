package main

import "testing"

func TestResolveAddr(t *testing.T) {
	got, err := resolveAddr(nil, "")
	if err != nil || got != "127.0.0.1:8081" {
		t.Fatalf("default=%q err=%v", got, err)
	}
	got, err = resolveAddr(nil, "127.0.0.1:9000")
	if err != nil || got != "127.0.0.1:9000" {
		t.Fatalf("env=%q err=%v", got, err)
	}
	got, err = resolveAddr([]string{"127.0.0.1:18081"}, "127.0.0.1:9000")
	if err != nil || got != "127.0.0.1:18081" {
		t.Fatalf("arg=%q err=%v", got, err)
	}
	if _, err := resolveAddr([]string{"a", "b"}, ""); err == nil {
		t.Fatal("expected usage error")
	}
}
