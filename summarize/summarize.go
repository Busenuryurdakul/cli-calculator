package summarize

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	ErrEmptyPath       = errors.New("empty path")
	ErrMissingField    = errors.New("missing required field")
	ErrInvalidMinCount = errors.New("invalid min_count")
	ErrEmptyInput      = errors.New("empty input")
	ErrTrailingJSON    = errors.New("trailing JSON")
)

const (
	defaultMinCount = 1
	scanBufSize     = 64 * 1024
	maxScanToken    = 1024 * 1024
)

// Config is loaded from a JSON file. Input and Output are required after trim.
// Relative input/output paths are resolved against the config file's directory.
type Config struct {
	Input    string `json:"input"`
	Output   string `json:"output"`
	MinCount *int   `json:"min_count,omitempty"`
}

// MinCountValue returns min_count, defaulting to 1 when omitted.
func (c Config) MinCountValue() int {
	if c.MinCount == nil {
		return defaultMinCount
	}
	return *c.MinCount
}

// Summary is the JSON report written after counting words.
// TotalWords is the unfiltered count; UniqueWords and Frequencies are after min_count.
type Summary struct {
	Source      string         `json:"source"`
	TotalWords  int            `json:"total_words"`
	UniqueWords int            `json:"unique_words"`
	Frequencies map[string]int `json:"frequencies"`
}

// LoadConfig reads JSON config, rejects unknown/trailing data, validates, and
// resolves relative paths against the config file directory.
func LoadConfig(path string) (cfg Config, err error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, fmt.Errorf("load config: %w", ErrEmptyPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, wrapFileError("load config", path, err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&cfg); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			if uerr := json.Unmarshal(data, new(json.RawMessage)); uerr != nil {
				return Config{}, fmt.Errorf("load config %q: %w", path, uerr)
			}
		}
		return Config{}, fmt.Errorf("load config %q: %w", path, err)
	}
	if err = rejectTrailing(dec); err != nil {
		return Config{}, fmt.Errorf("load config %q: %w", path, err)
	}
	if err = Validate(&cfg); err != nil {
		return Config{}, fmt.Errorf("load config %q: %w", path, err)
	}
	resolvePaths(path, &cfg)
	return cfg, nil
}

// Validate trims required fields and applies min_count default/rules.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("%w: config", ErrMissingField)
	}
	cfg.Input = strings.TrimSpace(cfg.Input)
	cfg.Output = strings.TrimSpace(cfg.Output)
	if cfg.Input == "" {
		return fmt.Errorf("%w: input", ErrMissingField)
	}
	if cfg.Output == "" {
		return fmt.Errorf("%w: output", ErrMissingField)
	}
	if cfg.MinCount == nil {
		n := defaultMinCount
		cfg.MinCount = &n
	} else if *cfg.MinCount < 1 {
		return fmt.Errorf("%w: must be >= 1", ErrInvalidMinCount)
	}
	return nil
}

func resolvePaths(configPath string, cfg *Config) {
	base := filepath.Dir(configPath)
	cfg.Input = resolvePath(base, cfg.Input)
	cfg.Output = resolvePath(base, cfg.Output)
}

func resolvePath(base, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(base, p))
}

func rejectTrailing(dec *json.Decoder) error {
	var extra json.RawMessage
	err := dec.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return ErrTrailingJSON
}

// CountWords tallies normalized word frequencies from any io.Reader.
func CountWords(r io.Reader) (map[string]int, error) {
	freq := make(map[string]int)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, scanBufSize), maxScanToken)
	scanner.Split(bufio.ScanWords)

	total := 0
	for scanner.Scan() {
		word := normalize(scanner.Text())
		if word == "" {
			continue
		}
		freq[word]++
		total++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("count words: %w", err)
	}
	if total == 0 {
		return nil, ErrEmptyInput
	}
	return freq, nil
}

// CountWordsFile reads a text file and counts word frequencies.
func CountWordsFile(path string) (map[string]int, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("count words: %w", ErrEmptyPath)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, wrapFileError("count words", path, err)
	}
	defer f.Close()

	freq, err := CountWords(f)
	if err != nil {
		return nil, fmt.Errorf("count words %q: %w", path, err)
	}
	return freq, nil
}

// BuildSummary keeps unfiltered TotalWords and min_count-filtered frequencies.
func BuildSummary(source string, freq map[string]int, minCount int) Summary {
	total := 0
	for _, n := range freq {
		total += n
	}

	filtered := make(map[string]int, len(freq))
	for word, n := range freq {
		if n < minCount {
			continue
		}
		filtered[word] = n
	}
	return Summary{
		Source:      source,
		TotalWords:  total,
		UniqueWords: len(filtered),
		Frequencies: filtered,
	}
}

// WriteSummary encodes indented JSON and checks encode plus close errors.
func WriteSummary(path string, summary Summary) (err error) {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("write summary: %w", ErrEmptyPath)
	}

	f, err := os.Create(path)
	if err != nil {
		return wrapFileError("write summary", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close summary %q: %w", path, closeErr)
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err = enc.Encode(summary); err != nil {
		return fmt.Errorf("write summary %q: %w", path, err)
	}
	return nil
}

// Run loads config, counts words, filters, and writes the JSON summary.
func Run(configPath string) (Summary, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return Summary{}, err
	}

	freq, err := CountWordsFile(cfg.Input)
	if err != nil {
		return Summary{}, err
	}

	summary := BuildSummary(cfg.Input, freq, cfg.MinCountValue())
	if err := WriteSummary(cfg.Output, summary); err != nil {
		return Summary{}, err
	}
	return summary, nil
}

func normalize(word string) string {
	word = strings.ToLower(word)
	return strings.TrimFunc(word, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
}

func wrapFileError(op, path string, err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("%s %q: file not found: %w", op, path, err)
	}
	return fmt.Errorf("%s %q: %w", op, path, err)
}

// ToolSummary documents how this package combines earlier skills.
func ToolSummary() string {
	return `Combined skills in this package:

- LoadConfig: JSON tags, required fields, unknown/trailing rejection, wrapped file errors
- Relative input/output paths resolve against the config file directory
- min_count defaults to 1; values < 1 are invalid
- CountWords: bufio.ScanWords + map[string]int, Unicode trim, long-token buffer
- total_words is unfiltered; frequencies/unique_words apply min_count after counting
- Run: load config, read data, write indented JSON; fail fast before creating output`
}
