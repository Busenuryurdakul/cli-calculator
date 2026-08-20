package payload

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

func object(data []byte) map[string]any {
	var raw []map[string]any
	if err := json.Unmarshal(data, &raw); err != nil || len(raw) != 1 {
		return nil
	}
	return raw[0]
}

func TestMarshalUsesStructTags(t *testing.T) {
	data, err := Marshal([]Reading{{
		Location: "Ankara",
		Celsius:  10,
		Source:   "sensor",
		Internal: "secret",
	}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	got := object(data)
	if got == nil {
		t.Fatalf("could not inspect JSON: %s", data)
	}
	if _, ok := got["Location"]; ok {
		t.Fatal("exported Go name Location should not appear; json tag is location")
	}
	if _, ok := got["Internal"]; ok {
		t.Fatal("json:\"-\" field Internal must be omitted")
	}
	if got["location"] != "Ankara" || got["celsius"] != 10.0 || got["source"] != "sensor" {
		t.Fatalf("unexpected JSON object: %#v", got)
	}
}

func TestOmitemptyNilOptionalFields(t *testing.T) {
	data, err := Marshal([]Reading{{Location: "Izmir", Celsius: 20}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := object(data)
	if _, ok := got["source"]; ok {
		t.Fatal("empty source should be omitted")
	}
	if _, ok := got["note"]; ok {
		t.Fatal("nil note should be omitted")
	}
	if _, ok := got["humidity"]; ok {
		t.Fatal("nil humidity should be omitted")
	}
}

func TestPointerZeroValuesAreEncoded(t *testing.T) {
	data, err := Marshal([]Reading{{
		Location: "Ankara",
		Celsius:  0,
		Note:     Note(""),
		Humidity: Float(0),
	}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := object(data)
	if _, ok := got["note"]; !ok {
		t.Fatal("pointer to empty string should be present")
	}
	if got["note"] != "" {
		t.Fatalf("note got %#v, want empty string", got["note"])
	}
	if _, ok := got["humidity"]; !ok {
		t.Fatal("pointer to 0 should be present")
	}
	if got["humidity"] != 0.0 {
		t.Fatalf("humidity got %#v, want 0", got["humidity"])
	}
}

func TestUnmarshalIgnoresUnknownFields(t *testing.T) {
	data := []byte(`[{"location":"Ankara","celsius":10,"source":"sensor","extra":true}]`)
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal should ignore unknown fields: %v", err)
	}
	if len(got) != 1 || got[0].Location != "Ankara" || got[0].Celsius != 10 || got[0].Source != "sensor" {
		t.Fatalf("unexpected payload: %+v", got)
	}
	if got[0].Note != nil {
		t.Fatal("missing note should stay nil")
	}
	if got[0].Internal != "" {
		t.Fatal("json:\"-\" field should not unmarshal from JSON")
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	_, err := Unmarshal([]byte(`{"location":`))
	if err == nil {
		t.Fatal("expected syntax error")
	}
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) {
		t.Fatalf("got %T, want wrapped json.SyntaxError", err)
	}
}

func TestUnmarshalTypeMismatch(t *testing.T) {
	_, err := Unmarshal([]byte(`[{"location":"Ankara","celsius":"hot"}]`))
	if err == nil {
		t.Fatal("expected type mismatch")
	}
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("got %T, want wrapped json.UnmarshalTypeError", err)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	want := []Reading{
		FromDomain(reading.TemperatureReading{Location: "Ankara", Celsius: 10, Source: "sensor"}, Note("hot")),
		FromDomain(reading.TemperatureReading{Location: "Van", Celsius: -2, Source: "station"}, nil),
	}

	var buf bytes.Buffer
	if err := Encode(&buf, want); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d readings, want 2", len(got))
	}
	if got[0].Location != "Ankara" || got[0].Note == nil || *got[0].Note != "hot" {
		t.Fatalf("first reading: %+v", got[0])
	}
	if got[1].Location != "Van" || got[1].Note != nil {
		t.Fatalf("second reading: %+v", got[1])
	}
	if ToDomain(got[0]).Source != "sensor" {
		t.Fatalf("ToDomain source got %q", ToDomain(got[0]).Source)
	}
}

func TestDecodeStrictRejectsUnknownFields(t *testing.T) {
	raw := `[{"location":"Ankara","celsius":10,"mystery":1}]`
	if _, err := Decode(strings.NewReader(raw)); err != nil {
		t.Fatalf("Decode should ignore unknown fields: %v", err)
	}

	_, err := DecodeStrict(strings.NewReader(raw))
	if err == nil {
		t.Fatal("expected unknown-field error")
	}
}

func TestDecodeRejectsTrailingJSON(t *testing.T) {
	_, err := Decode(strings.NewReader(`[{"location":"Ankara","celsius":10}][{"location":"Izmir","celsius":20}]`))
	if !errors.Is(err, ErrTrailingJSON) {
		t.Fatalf("got %v, want ErrTrailingJSON", err)
	}
}

func TestDecodeStreamMultipleAndBadIndex(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeStream(&buf, []Reading{
		{Location: "Ankara", Celsius: 10},
		{Location: "Izmir", Celsius: 20},
	}); err != nil {
		t.Fatalf("EncodeStream: %v", err)
	}

	got, err := DecodeStream(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeStream: %v", err)
	}
	if len(got) != 2 || got[0].Location != "Ankara" || got[1].Location != "Izmir" {
		t.Fatalf("unexpected stream: %+v", got)
	}

	_, err = DecodeStream(strings.NewReader("{\"location\":\"Ankara\",\"celsius\":10}\n{bad json\n"))
	if err == nil {
		t.Fatal("expected decode error on second record")
	}
	var streamErr *StreamError
	if !errors.As(err, &streamErr) || streamErr.Index != 1 {
		t.Fatalf("got %v, want StreamError index 1", err)
	}
}

func TestDecodeReaderError(t *testing.T) {
	want := errors.New("read failed")
	_, err := Decode(errReader{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
	}
}

func TestEncodeWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := Encode(errWriter{err: want}, []Reading{{Location: "Ankara", Celsius: 1}})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped %v", err, want)
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

var _ io.Reader = errReader{}
var _ io.Writer = errWriter{}
