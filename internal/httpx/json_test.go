package httpx

import (
	"strings"
	"testing"
)

func TestIsJSONContentType(t *testing.T) {
	if !IsJSONContentType("application/json") {
		t.Fatal("plain json")
	}
	if !IsJSONContentType("application/json; charset=utf-8") {
		t.Fatal("charset json")
	}
	if IsJSONContentType("text/plain") {
		t.Fatal("text/plain accepted")
	}
}

func TestDecodeJSONErrors(t *testing.T) {
	var dst map[string]any
	if err := DecodeJSON(strings.NewReader(""), &dst); err != ErrEmptyBody {
		t.Fatalf("empty: %v", err)
	}
	if err := DecodeJSON(strings.NewReader(`{"a":1}{}`), &dst); err != ErrTrailingJSON {
		t.Fatalf("trailing: %v", err)
	}
	if err := DecodeJSON(strings.NewReader(`{"a":`), &dst); err != ErrInvalidJSON {
		t.Fatalf("malformed: %v", err)
	}
	if err := DecodeJSON(strings.NewReader(`{"a":1,"nope":true}`), &struct {
		A int `json:"a"`
	}{}); err != ErrInvalidJSON {
		t.Fatalf("unknown field: %v", err)
	}
}
