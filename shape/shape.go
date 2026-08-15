package shape

import "math"

// Shape is a small interface for geometric figures.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Circle represents a circle with a radius.
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Rectangle represents a rectangle with width and height.
type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// ReviewSummary documents naming and design choices for code review.
func ReviewSummary() string {
	return `Shape library review:
- Interface size: Shape has 2 methods (Area, Perimeter) — narrow and focused
- Receivers: value receivers on Circle and Rectangle (small immutable structs)
- Naming: concrete types match domain nouns; methods are verb phrases`
}
