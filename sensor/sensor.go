package sensor

import (
	"fmt"
	"reflect"
)

// Measurer is a narrow interface: one method, many concrete types.
type Measurer interface {
	MeasureCelsius() float64
}

// Labeled is a second small interface for optional descriptions.
type Labeled interface {
	Label() string
}

// StationReading is an automated station sensor (implicit Measurer + Labeled).
type StationReading struct {
	Name    string
	Celsius float64
}

func (s StationReading) MeasureCelsius() float64 { return s.Celsius }
func (s StationReading) Label() string           { return s.Name }

// ManualGauge is a hand-entered reading (also satisfies Measurer implicitly).
type ManualGauge struct {
	Location string
	Celsius  float64
}

func (m ManualGauge) MeasureCelsius() float64 { return m.Celsius }
func (m ManualGauge) Label() string           { return m.Location + " (manual)" }

// Describe calls methods on any Measurer stored behind an interface value.
func Describe(m Measurer) string {
	if IsNil(m) {
		return "<nil measurer>"
	}

	reading := fmt.Sprintf("%.1f°C", m.MeasureCelsius())
	if labeled, ok := m.(Labeled); ok {
		return labeled.Label() + ": " + reading
	}
	return reading
}

// PrintReadings demonstrates polymorphism over a slice of interface values.
func PrintReadings(readings []Measurer) []string {
	lines := make([]string, 0, len(readings))
	for _, reading := range readings {
		lines = append(lines, Describe(reading))
	}
	return lines
}

// IsNil reports whether m is an untyped nil interface or an interface
// holding a nil concrete pointer/map/slice/etc.
func IsNil(m Measurer) bool {
	if m == nil {
		return true
	}

	value := reflect.ValueOf(m)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// SafeMeasure avoids calling methods on nil interface values.
func SafeMeasure(m Measurer) (float64, bool) {
	if IsNil(m) {
		return 0, false
	}
	return m.MeasureCelsius(), true
}

// NilInterfaceSummary documents the two nil cases.
func NilInterfaceSummary() string {
	return `Nil interface pitfalls:

1) var m Measurer           -> m == nil (no type, no value)
2) var p *StationReading    -> var m Measurer = p; m != nil but p is nil

Case 2 is subtle: the interface holds (type=*StationReading, value=nil).
Calling m.MeasureCelsius() panics. Use IsNil(m) or SafeMeasure(m) first.`
}
