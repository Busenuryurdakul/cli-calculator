package tempconv

import (
	"math"
	"testing"
)

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name string
		c    float64
		want float64
	}{
		{name: "freezing point", c: 0, want: 32},
		{name: "boiling point", c: 100, want: 212},
		{name: "body temperature", c: 37, want: 98.6},
		{name: "negative", c: -40, want: -40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CelsiusToFahrenheit(tt.c)
			if !almostEqual(got, tt.want, 1e-9) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFahrenheitToCelsius(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want float64
	}{
		{name: "freezing point", f: 32, want: 0},
		{name: "boiling point", f: 212, want: 100},
		{name: "negative forty", f: -40, want: -40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FahrenheitToCelsius(tt.f)
			if !almostEqual(got, tt.want, 1e-9) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := 25.5
	back := FahrenheitToCelsius(CelsiusToFahrenheit(original))
	if !almostEqual(back, original, 1e-9) {
		t.Fatalf("round trip got %v, want %v", back, original)
	}
}
