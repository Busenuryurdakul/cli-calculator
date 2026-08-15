package reading

import (
	"strings"
	"testing"
)

func TestPositionalLiteral(t *testing.T) {
	// Positional literal: field order must match the struct definition.
	r := TemperatureReading{"Bursa", 15.0, "manual"}
	if r.Location != "Bursa" {
		t.Fatalf("unexpected location: %q", r.Location)
	}
}

func TestNew(t *testing.T) {
	r := New("Ankara", 22.5, "sensor")
	if r.Location != "Ankara" || r.Celsius != 22.5 || r.Source != "sensor" {
		t.Fatalf("unexpected reading: %+v", r)
	}
}

func TestFahrenheit(t *testing.T) {
	r := TemperatureReading{Location: "test", Celsius: 100, Source: "manual"}
	if r.Fahrenheit() != 212 {
		t.Fatalf("got %v, want 212", r.Fahrenheit())
	}
}

func TestIsFreezing(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    bool
	}{
		{name: "below freezing", celsius: -5, want: true},
		{name: "at freezing", celsius: 0, want: true},
		{name: "above freezing", celsius: 10, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := TemperatureReading{Celsius: tt.celsius}
			if got := r.IsFreezing(); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateCelsius(t *testing.T) {
	r := New("Izmir", 10, "station")
	r.UpdateCelsius(25)
	if r.Celsius != 25 {
		t.Fatalf("got %v, want 25", r.Celsius)
	}
}

func TestMeasureCelsius(t *testing.T) {
	r := TemperatureReading{Celsius: 21}
	if r.MeasureCelsius() != 21 {
		t.Fatalf("got %v, want 21", r.MeasureCelsius())
	}
}

func TestString(t *testing.T) {
	r := TemperatureReading{Location: "Istanbul", Celsius: 0, Source: "manual"}
	got := r.String()
	if !strings.Contains(got, "Istanbul") || !strings.Contains(got, "32.0°F") {
		t.Fatalf("unexpected string: %q", got)
	}
}
