package report

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

func sample(location string, celsius float64, source string) reading.TemperatureReading {
	return reading.TemperatureReading{Location: location, Celsius: celsius, Source: source}
}

func samples() []reading.TemperatureReading {
	return []reading.TemperatureReading{
		sample("Ankara", 10, "sensor"),
		sample("Izmir", 20, "station"),
	}
}

func equalReadings(t *testing.T, got, want []reading.TemperatureReading) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d readings, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("reading[%d]=%+v, want %+v", i, got[i], want[i])
		}
	}
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) {
	return 0, r.err
}

type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return 1, io.ErrShortWrite
}

func TestWriteAllAndReadAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readings.txt")
	want := samples()

	if err := WriteAll(path, want); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	got, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	equalReadings(t, got, want)
}

func TestWriteBufferedAndReadLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readings.txt")
	want := samples()

	if err := WriteBuffered(path, want); err != nil {
		t.Fatalf("WriteBuffered: %v", err)
	}

	got, err := ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	equalReadings(t, got, want)
}

func TestReadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")

	_, err := ReadAll(path)
	if err == nil {
		t.Fatal("expected missing-file error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("wrapped error should match os.ErrNotExist, got %v", err)
	}

	_, err = ReadLines(path)
	if err == nil {
		t.Fatal("expected missing-file error from ReadLines")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadLines wrapped error should match os.ErrNotExist, got %v", err)
	}
}

func TestEmptyPath(t *testing.T) {
	if err := WriteAll("", samples()); !errors.Is(err, ErrEmptyPath) {
		t.Fatalf("WriteAll empty path: %v", err)
	}
	if _, err := ReadAll(""); !errors.Is(err, ErrEmptyPath) {
		t.Fatalf("ReadAll empty path: %v", err)
	}
}

func TestCopyAndCountLinesWithBuffers(t *testing.T) {
	var src bytes.Buffer
	if err := Format(&src, samples()); err != nil {
		t.Fatalf("Format: %v", err)
	}

	var dst bytes.Buffer
	n, err := Copy(&dst, strings.NewReader(src.String()))
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if n == 0 || dst.String() != src.String() {
		t.Fatalf("copy mismatch: n=%d dst=%q src=%q", n, dst.String(), src.String())
	}

	count, err := CountLines(strings.NewReader(dst.String()))
	if err != nil {
		t.Fatalf("CountLines: %v", err)
	}
	if count != 2 {
		t.Fatalf("CountLines got %d, want 2", count)
	}
}

func TestCopyReaderError(t *testing.T) {
	want := errors.New("read failed")
	n, err := Copy(&bytes.Buffer{}, errReader{err: want})
	if n != 0 {
		t.Fatalf("copied %d bytes, want 0", n)
	}
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}

func TestCountLinesReaderError(t *testing.T) {
	want := errors.New("read failed")
	_, err := CountLines(errReader{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}

func TestFormatWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := Format(errWriter{err: want}, samples())
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}

func TestFormatShortWrite(t *testing.T) {
	err := Format(shortWriter{}, samples())
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("got %v, want wrapped io.ErrShortWrite", err)
	}
}

func TestCopyShortWrite(t *testing.T) {
	src := strings.NewReader("Ankara\t10.0\tsensor\n")
	n, err := Copy(shortWriter{}, src)
	if err == nil {
		t.Fatal("expected short-write error")
	}
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("got %v, want wrapped io.ErrShortWrite", err)
	}
	if n < 0 {
		t.Fatalf("byte count should be non-negative, got %d", n)
	}
}

func TestParseLineError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(path, []byte("not-a-report\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := ReadLines(path)
	if !errors.Is(err, ErrInvalidLine) {
		t.Fatalf("got %v, want ErrInvalidLine", err)
	}
}

func TestReadLinesLongLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "long.txt")
	location := strings.Repeat("A", bufioDefaultToken+1024)
	line := location + "\t10.0\tsensor\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines long line: %v", err)
	}
	if len(got) != 1 || got[0].Location != location || got[0].Celsius != 10 {
		t.Fatalf("unexpected long-line reading: %+v", got)
	}
}

// bufioDefaultToken is the default Scanner token limit; our scanners allow more.
const bufioDefaultToken = 64 * 1024
