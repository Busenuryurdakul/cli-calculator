package report

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

var (
	ErrEmptyPath   = errors.New("empty path")
	ErrInvalidLine = errors.New("invalid report line")
)

const (
	scanBufSize  = 64 * 1024
	maxScanToken = 1024 * 1024
)

// Format writes readings as tab-separated lines to any io.Writer.
func Format(w io.Writer, items []reading.TemperatureReading) error {
	for _, item := range items {
		_, err := fmt.Fprintf(w, "%s\t%.1f\t%s\n", item.Location, item.Celsius, item.Source)
		if err != nil {
			return fmt.Errorf("format report: %w", err)
		}
	}
	return nil
}

// WriteAll creates a small report with os.WriteFile.
func WriteAll(path string, items []reading.TemperatureReading) error {
	if path == "" {
		return fmt.Errorf("write report: %w", ErrEmptyPath)
	}

	var buf bytes.Buffer
	if err := Format(&buf, items); err != nil {
		return fmt.Errorf("write report %q: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write report %q: %w", path, err)
	}
	return nil
}

// WriteBuffered creates a file and writes through a bufio.Writer.
func WriteBuffered(path string, items []reading.TemperatureReading) (err error) {
	if path == "" {
		return fmt.Errorf("write report: %w", ErrEmptyPath)
	}

	f, err := os.Create(path)
	if err != nil {
		return wrapFileError("create", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close report %q: %w", path, closeErr)
		}
	}()

	bw := bufio.NewWriter(f)
	if err = Format(bw, items); err != nil {
		return fmt.Errorf("write report %q: %w", path, err)
	}
	if err = bw.Flush(); err != nil {
		return fmt.Errorf("flush report %q: %w", path, err)
	}
	return nil
}

// ReadAll loads a whole file with os.ReadFile, then parses lines.
func ReadAll(path string) ([]reading.TemperatureReading, error) {
	if path == "" {
		return nil, fmt.Errorf("read report: %w", ErrEmptyPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, wrapFileError("read", path, err)
	}
	return parseReader(bytes.NewReader(data), path)
}

// ReadLines opens a file and scans it line by line with bufio.Scanner.
func ReadLines(path string) ([]reading.TemperatureReading, error) {
	if path == "" {
		return nil, fmt.Errorf("read report: %w", ErrEmptyPath)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, wrapFileError("read", path, err)
	}
	defer f.Close()

	return parseReader(f, path)
}

// Copy copies from any io.Reader to any io.Writer.
func Copy(dst io.Writer, src io.Reader) (int64, error) {
	n, err := io.Copy(dst, src)
	if err != nil {
		return n, fmt.Errorf("copy report: %w", err)
	}
	return n, nil
}

// CountLines counts lines from any io.Reader using a buffered scanner.
func CountLines(r io.Reader) (int, error) {
	scanner := newScanner(r)
	n := 0
	for scanner.Scan() {
		n++
	}
	if err := scanner.Err(); err != nil {
		return n, fmt.Errorf("count lines: %w", err)
	}
	return n, nil
}

func newScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, scanBufSize), maxScanToken)
	return scanner
}

func parseReader(r io.Reader, path string) ([]reading.TemperatureReading, error) {
	var items []reading.TemperatureReading
	scanner := newScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		item, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("read report %q: %w", path, err)
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read report %q: %w", path, err)
	}
	return items, nil
}

func parseLine(line string) (reading.TemperatureReading, error) {
	parts := strings.Split(line, "\t")
	if len(parts) != 3 {
		return reading.TemperatureReading{}, fmt.Errorf("%w: %q", ErrInvalidLine, line)
	}
	celsius, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return reading.TemperatureReading{}, fmt.Errorf("%w: celsius %q: %w", ErrInvalidLine, parts[1], err)
	}
	return reading.TemperatureReading{
		Location: parts[0],
		Celsius:  celsius,
		Source:   parts[2],
	}, nil
}

func wrapFileError(op, path string, err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("%s report %q: file not found: %w", op, path, err)
	}
	return fmt.Errorf("%s report %q: %w", op, path, err)
}

// FileSummary documents the file I/O patterns used here.
func FileSummary() string {
	return `File I/O patterns in this package:

- os.ReadFile loads a small file at once; bufio.Scanner reads line by line
- os.WriteFile writes a complete buffer; bufio.Writer batches larger output
- Format and Copy depend on io.Reader / io.Writer, not concrete file types
- os.IsNotExist is checked at the file-system boundary, then errors are wrapped with %w`
}
