package summarize

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", name, err)
	}
	return path
}

func writeConfig(t *testing.T, dir, body string) string {
	t.Helper()
	return writeFile(t, dir, "config.json", body)
}

func TestLoadConfigValid(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `{"input":"in.txt","output":"out.json","min_count":2}`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Input != filepath.Join(dir, "in.txt") || cfg.Output != filepath.Join(dir, "out.json") {
		t.Fatalf("paths should resolve against config dir, got input=%q output=%q", cfg.Input, cfg.Output)
	}
	if cfg.MinCountValue() != 2 {
		t.Fatalf("min_count got %d, want 2", cfg.MinCountValue())
	}
}

func TestLoadConfigDefaultMinCount(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `{"input":"in.txt","output":"out.json"}`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.MinCountValue() != 1 {
		t.Fatalf("default min_count got %d, want 1", cfg.MinCountValue())
	}
}

func TestLoadConfigTrimsRequiredFields(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `{"input":"  in.txt  ","output":"  out.json  "}`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Input != filepath.Join(dir, "in.txt") || cfg.Output != filepath.Join(dir, "out.json") {
		t.Fatalf("trimmed paths got input=%q output=%q", cfg.Input, cfg.Output)
	}
}

func TestLoadConfigMissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `{"input":"   ","output":"out.json"}`)
	_, err := LoadConfig(path)
	if !errors.Is(err, ErrMissingField) {
		t.Fatalf("got %v, want ErrMissingField", err)
	}
}

func TestLoadConfigInvalidMinCount(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `{"input":"in.txt","output":"out.json","min_count":0}`)
	_, err := LoadConfig(path)
	if !errors.Is(err, ErrInvalidMinCount) {
		t.Fatalf("got %v, want ErrInvalidMinCount", err)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, want os.ErrNotExist", err)
	}
}

func TestLoadConfigMalformedJSON(t *testing.T) {
	path := writeFile(t, t.TempDir(), "bad.json", `{"input":`)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) {
		t.Fatalf("got %T, want wrapped json.SyntaxError", err)
	}
}

func TestLoadConfigUnknownField(t *testing.T) {
	path := writeConfig(t, t.TempDir(), `{"input":"in.txt","output":"out.json","typo":true}`)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected unknown-field error")
	}
}

func TestLoadConfigTrailingJSON(t *testing.T) {
	path := writeFile(t, t.TempDir(), "trail.json", `{"input":"in.txt","output":"out.json"}{"x":1}`)
	_, err := LoadConfig(path)
	if !errors.Is(err, ErrTrailingJSON) {
		t.Fatalf("got %v, want ErrTrailingJSON", err)
	}
}

func TestCountWordsCaseInsensitive(t *testing.T) {
	freq, err := CountWords(strings.NewReader("Go go, maps; maps maps."))
	if err != nil {
		t.Fatalf("CountWords: %v", err)
	}
	if freq["go"] != 2 || freq["maps"] != 3 {
		t.Fatalf("unexpected frequencies: %#v", freq)
	}
}

func TestCountWordsPunctuationAndUnicode(t *testing.T) {
	freq, err := CountWords(strings.NewReader(`"Café", naïve! don't.`))
	if err != nil {
		t.Fatalf("CountWords: %v", err)
	}
	if freq["café"] != 1 || freq["naïve"] != 1 || freq["don't"] != 1 {
		t.Fatalf("unexpected unicode/punct frequencies: %#v", freq)
	}
}

func TestCountWordsEmptyAndPunctuationOnly(t *testing.T) {
	_, err := CountWords(strings.NewReader("   \n\t"))
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("blank input: got %v, want ErrEmptyInput", err)
	}
	_, err = CountWords(strings.NewReader("... !!! ---"))
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("punctuation-only: got %v, want ErrEmptyInput", err)
	}
}

func TestCountWordsLongToken(t *testing.T) {
	token := strings.Repeat("a", 64*1024+2048)
	freq, err := CountWords(strings.NewReader(token + "\n"))
	if err != nil {
		t.Fatalf("long token: %v", err)
	}
	if freq[token] != 1 || utf8.RuneCountInString(token) < 64*1024 {
		t.Fatalf("long token not counted: %#v", freq)
	}
}

func TestCountWordsFileMissing(t *testing.T) {
	_, err := CountWordsFile(filepath.Join(t.TempDir(), "missing.txt"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, want os.ErrNotExist", err)
	}
}

func TestCountWordsReaderError(t *testing.T) {
	want := errors.New("read failed")
	_, err := CountWords(errReader{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}

func TestBuildSummaryKeepsUnfilteredTotal(t *testing.T) {
	summary := BuildSummary("in.txt", map[string]int{"a": 1, "b": 3}, 2)
	if _, ok := summary.Frequencies["a"]; ok {
		t.Fatalf("word a should be filtered: %+v", summary)
	}
	if summary.Frequencies["b"] != 3 {
		t.Fatalf("filtered frequency: %+v", summary)
	}
	if summary.TotalWords != 4 {
		t.Fatalf("total_words should stay unfiltered, got %d", summary.TotalWords)
	}
	if summary.UniqueWords != 1 {
		t.Fatalf("unique_words should be filtered size, got %d", summary.UniqueWords)
	}
}

func TestRunHappyPathTempDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "notes.txt", "red blue red")
	config := writeConfig(t, dir, `{"input":"notes.txt","output":"summary.json","min_count":1}`)

	got, err := Run(config)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got.TotalWords != 3 || got.UniqueWords != 2 || got.Frequencies["red"] != 2 || got.Frequencies["blue"] != 1 {
		t.Fatalf("unexpected summary: %+v", got)
	}
	if got.Source != filepath.Join(dir, "notes.txt") {
		t.Fatalf("source should be resolved path, got %q", got.Source)
	}

	data, err := os.ReadFile(filepath.Join(dir, "summary.json"))
	if err != nil {
		t.Fatalf("ReadFile summary: %v", err)
	}
	var fromDisk Summary
	if err := json.Unmarshal(data, &fromDisk); err != nil {
		t.Fatalf("unmarshal written summary: %v", err)
	}
	if fromDisk.TotalWords != 3 || fromDisk.UniqueWords != 2 || fromDisk.Frequencies["red"] != 2 {
		t.Fatalf("written summary mismatch: %+v", fromDisk)
	}
	if !strings.Contains(string(data), "\n  ") {
		t.Fatal("summary JSON should be indented")
	}
}

func TestRunDoesNotCreateOutputOnInputError(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "summary.json")
	config := writeConfig(t, dir, `{"input":"missing.txt","output":"summary.json"}`)

	_, err := Run(config)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, want os.ErrNotExist", err)
	}
	if _, statErr := os.Stat(output); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("output file must not be created when input is missing")
	}
}

func TestRunEmptyInputFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "empty.txt", "\n")
	output := filepath.Join(dir, "summary.json")
	config := writeConfig(t, dir, `{"input":"empty.txt","output":"summary.json"}`)

	_, err := Run(config)
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("got %v, want ErrEmptyInput", err)
	}
	if _, statErr := os.Stat(output); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("output file must not be created when input is empty")
	}
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) {
	return 0, r.err
}
