package shape

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

func TestShapeAreaAndPerimeter(t *testing.T) {
	tests := []struct {
		name      string
		shape     Shape
		wantArea  float64
		wantPerim float64
	}{
		{
			name:      "circle radius 3",
			shape:     Circle{Radius: 3},
			wantArea:  math.Pi * 9,
			wantPerim: 2 * math.Pi * 3,
		},
		{
			name:      "rectangle 4x5",
			shape:     Rectangle{Width: 4, Height: 5},
			wantArea:  20,
			wantPerim: 18,
		},
		{
			name:      "circle radius 0",
			shape:     Circle{Radius: 0},
			wantArea:  0,
			wantPerim: 0,
		},
		{
			name:      "rectangle unit square",
			shape:     Rectangle{Width: 1, Height: 1},
			wantArea:  1,
			wantPerim: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArea := tt.shape.Area()
			gotPerim := tt.shape.Perimeter()

			if !almostEqual(gotArea, tt.wantArea) {
				t.Fatalf("Area() = %v, want %v", gotArea, tt.wantArea)
			}
			if !almostEqual(gotPerim, tt.wantPerim) {
				t.Fatalf("Perimeter() = %v, want %v", gotPerim, tt.wantPerim)
			}
		})
	}
}

func TestShapeInterface(t *testing.T) {
	shapes := []Shape{
		Circle{Radius: 2},
		Rectangle{Width: 3, Height: 4},
	}

	if len(shapes) != 2 {
		t.Fatalf("expected 2 shapes, got %d", len(shapes))
	}
}
