package pipeline

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

var (
	ErrEmptyLocation     = errors.New("empty location")
	ErrEmptySource       = errors.New("empty source")
	ErrEmptyCelsius      = errors.New("empty celsius")
	ErrBelowAbsoluteZero = errors.New("temperature below absolute zero")
	ErrInvalidUnit       = errors.New("invalid unit")
)

const absoluteZeroC = -273.15

// ParseError is a typed error so callers can inspect it with errors.As.
type ParseError struct {
	Input string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse %q: %v", e.Input, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Result is a converted reading produced by Process.
type Result struct {
	Reading   reading.TemperatureReading
	Converted float64
	Unit      string
}

type Kind string

const (
	KindParse      Kind = "parse"
	KindValidation Kind = "validation"
	KindConvert    Kind = "convert"
	KindUnknown    Kind = "unknown"
)

// ParseCelsius returns (float64, error) and wraps syntax failures.
func ParseCelsius(s string) (float64, error) {
	if s == "" {
		return 0, &ParseError{Input: s, Err: ErrEmptyCelsius}
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, &ParseError{Input: s, Err: fmt.Errorf("not a number: %w", err)}
	}
	return n, nil
}

// ParseReading builds a TemperatureReading, failing fast on the first bad field.
func ParseReading(location, celsiusStr, source string) (reading.TemperatureReading, error) {
	if location == "" {
		return reading.TemperatureReading{}, ErrEmptyLocation
	}
	if source == "" {
		return reading.TemperatureReading{}, ErrEmptySource
	}

	celsius, err := ParseCelsius(celsiusStr)
	if err != nil {
		return reading.TemperatureReading{}, fmt.Errorf("reading %s: %w", location, err)
	}

	return reading.TemperatureReading{
		Location: location,
		Celsius:  celsius,
		Source:   source,
	}, nil
}

// Validate checks domain rules after a reading has been parsed.
func Validate(r reading.TemperatureReading) error {
	if r.Location == "" {
		return fmt.Errorf("validate: %w", ErrEmptyLocation)
	}
	if r.Source == "" {
		return fmt.Errorf("validate: %w", ErrEmptySource)
	}
	if r.Celsius < absoluteZeroC {
		return fmt.Errorf("validate %s: %w (%.2f°C)", r.Location, ErrBelowAbsoluteZero, r.Celsius)
	}
	return nil
}

// Convert returns the reading in C or F.
func Convert(r reading.TemperatureReading, unit string) (float64, error) {
	switch strings.ToUpper(unit) {
	case "C", "CELSIUS":
		return r.Celsius, nil
	case "F", "FAHRENHEIT":
		return r.Fahrenheit(), nil
	default:
		if unit == "" {
			return 0, fmt.Errorf("%w: empty (use C or F)", ErrInvalidUnit)
		}
		return 0, fmt.Errorf("%w: %q (use C or F)", ErrInvalidUnit, unit)
	}
}

// Process orchestrates parse → validate → convert with early returns.
// Nested success blocks are avoided so a failure cannot leave partial state.
func Process(location, celsiusStr, source, unit string) (Result, error) {
	if unit == "" {
		unit = "C"
	}

	r, err := ParseReading(location, celsiusStr, source)
	if err != nil {
		return Result{}, fmt.Errorf("process: %w", err)
	}

	if err := Validate(r); err != nil {
		return Result{}, fmt.Errorf("process: %w", err)
	}

	converted, err := Convert(r, unit)
	if err != nil {
		return Result{}, fmt.Errorf("process: %w", err)
	}

	return Result{Reading: r, Converted: converted, Unit: unit}, nil
}

// Classify inspects an error chain with errors.As and errors.Is.
func Classify(err error) Kind {
	if err == nil {
		return KindUnknown
	}

	var parseErr *ParseError
	if errors.As(err, &parseErr) {
		return KindParse
	}

	switch {
	case errors.Is(err, ErrEmptyLocation),
		errors.Is(err, ErrEmptySource),
		errors.Is(err, ErrEmptyCelsius),
		errors.Is(err, ErrBelowAbsoluteZero):
		return KindValidation
	case errors.Is(err, ErrInvalidUnit):
		return KindConvert
	default:
		return KindUnknown
	}
}

// ErrorSummary documents the error-handling patterns used here.
func ErrorSummary() string {
	return `Go error patterns in this package:

- Return (T, error) from ParseCelsius, ParseReading, Convert, Process
- Sentinel values via errors.New: ErrEmptyLocation, ErrBelowAbsoluteZero, ...
- fmt.Errorf for messages; %w to wrap and keep the chain
- errors.Is matches sentinels through wrapping
- errors.As extracts *ParseError from the chain
- Process fails fast: return on first error, no nested success blocks`
}
