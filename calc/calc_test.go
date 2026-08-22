package calc

import (
	"errors"
	"fmt"
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
		{name: "zero value", input: "0", want: 0},
		{name: "scientific", input: "1e2", want: 100},
		{name: "empty input", input: "", wantErr: true},
		{name: "whitespace only", input: " ", wantErr: true},
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
		{name: "zeros add", a: 0, b: 0, operator: "+", want: 0},
		{name: "zero dividend", a: 0, b: 5, operator: "/", want: 0},
		{name: "division by zero", a: 10, b: 0, operator: "/", wantErr: ErrDivisionByZero},
		{name: "empty operator", a: 1, b: 2, operator: "", wantErr: ErrInvalidOp},
		{name: "whitespace operator", a: 1, b: 2, operator: " + ", wantErr: ErrInvalidOp},
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

func TestParseNumberEmptySentinel(t *testing.T) {
	_, err := ParseNumber("")
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("got %v, want ErrEmptyInput", err)
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

func ExampleCalculate() {
	got, err := Calculate(10, 5, "+")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(got)
	// Output: 15
}

func ExampleParseNumber() {
	got, err := ParseNumber("3.5")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(got)
	// Output: 3.5
}

var (
	benchResult float64
	benchErr    error
)

func BenchmarkCalculate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var result float64
	var err error
	for i := 0; i < b.N; i++ {
		result, err = Calculate(10, 5, "+")
		if err != nil {
			b.Fatal(err)
		}
	}
	benchResult = result
	benchErr = err
}

func BenchmarkParseNumber(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var result float64
	var err error
	for i := 0; i < b.N; i++ {
		result, err = ParseNumber("42.5")
		if err != nil {
			b.Fatal(err)
		}
	}
	benchResult = result
	benchErr = err
}
