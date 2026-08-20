package payload

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

var ErrTrailingJSON = errors.New("trailing JSON")

// StreamError wraps a per-record encode/decode failure with its index.
type StreamError struct {
	Index int
	Err   error
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("decode reading %d: %v", e.Index, e.Err)
}

func (e *StreamError) Unwrap() error {
	return e.Err
}

// Reading is the JSON-facing view of a temperature measurement.
type Reading struct {
	Location string   `json:"location"`
	Celsius  float64  `json:"celsius"`
	Source   string   `json:"source,omitempty"`
	Note     *string  `json:"note,omitempty"`
	Humidity *float64 `json:"humidity,omitempty"`
	Internal string   `json:"-"`
}

// Note returns a pointer so optional JSON string fields can be set.
func Note(s string) *string {
	return &s
}

// Float returns a pointer so optional numeric JSON fields can be set.
func Float(v float64) *float64 {
	return &v
}

// FromDomain maps a domain reading into the JSON payload shape.
func FromDomain(r reading.TemperatureReading, note *string) Reading {
	return Reading{
		Location: r.Location,
		Celsius:  r.Celsius,
		Source:   r.Source,
		Note:     note,
	}
}

// ToDomain maps a JSON payload back into the domain type.
func ToDomain(r Reading) reading.TemperatureReading {
	return reading.TemperatureReading{
		Location: r.Location,
		Celsius:  r.Celsius,
		Source:   r.Source,
	}
}

// Marshal converts readings to JSON using struct tags.
func Marshal(items []Reading) ([]byte, error) {
	data, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal readings: %w", err)
	}
	return data, nil
}

// Unmarshal parses JSON into readings. Unknown fields are ignored.
func Unmarshal(data []byte) ([]Reading, error) {
	var items []Reading
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("unmarshal readings: %w", err)
	}
	return items, nil
}

// Encode streams one JSON value to any io.Writer.
func Encode(w io.Writer, items []Reading) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		return fmt.Errorf("encode readings: %w", err)
	}
	return nil
}

// Decode streams one JSON value from any io.Reader. Unknown fields are ignored.
func Decode(r io.Reader) ([]Reading, error) {
	return decodeList(r, false)
}

// DecodeStrict streams JSON and rejects objects with unknown fields.
func DecodeStrict(r io.Reader) ([]Reading, error) {
	return decodeList(r, true)
}

// EncodeStream writes each reading as its own JSON value (NDJSON).
func EncodeStream(w io.Writer, items []Reading) error {
	enc := json.NewEncoder(w)
	for i, item := range items {
		if err := enc.Encode(item); err != nil {
			return fmt.Errorf("encode reading %d: %w", i, err)
		}
	}
	return nil
}

// DecodeStream reads successive JSON values until EOF.
func DecodeStream(r io.Reader) ([]Reading, error) {
	dec := json.NewDecoder(r)
	var items []Reading
	for i := 0; ; i++ {
		var item Reading
		if err := dec.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				return items, nil
			}
			return nil, &StreamError{Index: i, Err: err}
		}
		items = append(items, item)
	}
}

func decodeList(r io.Reader, strict bool) ([]Reading, error) {
	dec := json.NewDecoder(r)
	if strict {
		dec.DisallowUnknownFields()
	}

	var items []Reading
	if err := dec.Decode(&items); err != nil {
		return nil, fmt.Errorf("decode readings: %w", err)
	}
	if err := rejectTrailing(dec); err != nil {
		return nil, err
	}
	return items, nil
}

func rejectTrailing(dec *json.Decoder) error {
	var extra json.RawMessage
	err := dec.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("decode readings: %w", ErrTrailingJSON)
	}
	return fmt.Errorf("decode readings: %w", ErrTrailingJSON)
}

// JSONSummary documents the encoding/json patterns used here.
func JSONSummary() string {
	return `JSON patterns in this package:

- json tags rename fields (location, celsius) and omit empty source/note/humidity
- Internal uses json:"-" and never appears in output
- json.Marshal / json.Unmarshal convert whole values in memory
- json.Encoder / json.Decoder stream through io.Writer / io.Reader
- unknown fields are ignored by default; DecodeStrict uses DisallowUnknownFields
- optional note/humidity are pointers: nil omits the field, a pointer includes it
- Decode rejects a second JSON value after the first (trailing JSON)`
}
