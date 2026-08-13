package calc

import (
	"errors"
	"math"
	"testing"
)

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{name: "valid integer", input: "42", want: 42},
		{name: "valid decimal", input: "3.14", want: 3.14},
		{name: "negative number", input: "-7.5", want: -7.5},
		{name: "empty input", input: "", wantErr: true},
		{name: "invalid text", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNumber(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		operator string
		want     float64
		wantErr  error
	}{
		{name: "addition", a: 10, b: 5, operator: "+", want: 15},
		{name: "subtraction", a: 10, b: 3, operator: "-", want: 7},
		{name: "multiplication", a: 4, b: 7, operator: "*", want: 28},
		{name: "division", a: 20, b: 4, operator: "/", want: 5},
		{name: "division by zero", a: 10, b: 0, operator: "/", wantErr: ErrDivisionByZero},
		{name: "invalid operator", a: 1, b: 2, operator: "%", wantErr: ErrInvalidOp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.a, tt.b, tt.operator)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateFloatingPoint(t *testing.T) {
	got, err := Calculate(0.1, 0.2, "+")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got-0.3) > 1e-9 {
		t.Fatalf("got %v, want ~0.3", got)
	}
}
