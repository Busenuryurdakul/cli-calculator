package pipeline

import (
	"errors"
	"strconv"
	"testing"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

func TestParseCelsius(t *testing.T) {
	got, err := ParseCelsius("22.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 22.5 {
		t.Fatalf("got %v, want 22.5", got)
	}
}

func TestParseCelsiusTypedError(t *testing.T) {
	_, err := ParseCelsius("abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if parseErr.Input != "abc" {
		t.Fatalf("input got %q", parseErr.Input)
	}
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Fatalf("expected strconv.ErrSyntax in chain, got %v", err)
	}
}

func TestParseCelsiusEmpty(t *testing.T) {
	_, err := ParseCelsius("")
	if !errors.Is(err, ErrEmptyCelsius) {
		t.Fatalf("got %v, want ErrEmptyCelsius", err)
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatal("empty input should still be a *ParseError")
	}
}

func TestParseReadingFailFast(t *testing.T) {
	_, err := ParseReading("", "not-a-number", "")
	if !errors.Is(err, ErrEmptyLocation) {
		t.Fatalf("first failure should be empty location, got %v", err)
	}
	if errors.As(err, new(*ParseError)) {
		t.Fatal("should not parse celsius after an earlier failure")
	}
}

func TestValidate(t *testing.T) {
	err := Validate(reading.TemperatureReading{Location: "Mars", Celsius: -400, Source: "probe"})
	if !errors.Is(err, ErrBelowAbsoluteZero) {
		t.Fatalf("got %v, want ErrBelowAbsoluteZero", err)
	}
}

func TestConvertInvalidUnit(t *testing.T) {
	_, err := Convert(reading.TemperatureReading{Location: "Izmir", Celsius: 18, Source: "manual"}, "K")
	if !errors.Is(err, ErrInvalidUnit) {
		t.Fatalf("got %v, want ErrInvalidUnit", err)
	}
}

func TestProcessSuccess(t *testing.T) {
	got, err := Process("Ankara", "22.5", "sensor", "F")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Converted != 72.5 {
		t.Fatalf("converted got %v, want 72.5", got.Converted)
	}
}

func TestProcessWrapsParseError(t *testing.T) {
	_, err := Process("Ankara", "abc", "sensor", "C")
	if err == nil {
		t.Fatal("expected error")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("wrapped error should still unwrap to *ParseError, got %v", err)
	}
	if parseErr.Input != "abc" {
		t.Fatalf("input got %q", parseErr.Input)
	}
	if Classify(err) != KindParse {
		t.Fatalf("classify got %s, want parse", Classify(err))
	}
}

func TestProcessWrapsSentinel(t *testing.T) {
	_, err := Process("Space", "-400", "probe", "C")
	if !errors.Is(err, ErrBelowAbsoluteZero) {
		t.Fatalf("wrapped error should still match sentinel, got %v", err)
	}
	if Classify(err) != KindValidation {
		t.Fatalf("classify got %s, want validation", Classify(err))
	}
}

func TestProcessFailFastStopsAtFirstError(t *testing.T) {
	_, err := Process("Ankara", "nope", "sensor", "K")
	if Classify(err) != KindParse {
		t.Fatalf("got %s, want parse (fail fast before convert)", Classify(err))
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Kind
	}{
		{name: "nil", err: nil, want: KindUnknown},
		{name: "parse", err: &ParseError{Input: "x", Err: ErrEmptyCelsius}, want: KindParse},
		{name: "empty location", err: ErrEmptyLocation, want: KindValidation},
		{name: "invalid unit", err: ErrInvalidUnit, want: KindConvert},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.err); got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}
