package reading

import (
	"fmt"

	"github.com/Busenuryurdakul/cli-calculator/tempconv"
)

// TemperatureReading represents a temperature measurement at a location.
type TemperatureReading struct {
	Location string
	Celsius  float64
	Source   string
}

// FreezingChecker is satisfied by both TemperatureReading and *TemperatureReading
// because IsFreezing uses a value receiver.
type FreezingChecker interface {
	IsFreezing() bool
}

// CelsiusUpdater is satisfied only by *TemperatureReading because UpdateCelsius
// uses a pointer receiver.
type CelsiusUpdater interface {
	UpdateCelsius(float64)
}

// DemoPositional returns a struct initialized with a positional literal.
func DemoPositional() TemperatureReading {
	return TemperatureReading{Location: "Izmir", Celsius: 25.0, Source: "station"}
}

// New returns a pointer initialized with keyed fields.
func New(location string, celsius float64, source string) *TemperatureReading {
	return &TemperatureReading{
		Location: location,
		Celsius:  celsius,
		Source:   source,
	}
}

// --- Value receivers (read-only, copy semantics) ---

// Fahrenheit converts the stored Celsius value without mutating the receiver.
func (r TemperatureReading) Fahrenheit() float64 {
	return tempconv.CelsiusToFahrenheit(r.Celsius)
}

// IsFreezing reports whether the reading is at or below freezing.
func (r TemperatureReading) IsFreezing() bool {
	return r.Celsius <= 0
}

// ScaledCopy returns a new reading with Celsius multiplied by factor.
// The receiver is copied, so the original value stays unchanged.
func (r TemperatureReading) ScaledCopy(factor float64) TemperatureReading {
	r.Celsius *= factor
	return r
}

// String implements fmt.Stringer for custom formatting.
func (r TemperatureReading) String() string {
	return fmt.Sprintf("%s: %.1f°C (%.1f°F) via %s", r.Location, r.Celsius, r.Fahrenheit(), r.Source)
}

// MeasureCelsius lets TemperatureReading satisfy sensor.Measurer implicitly.
func (r TemperatureReading) MeasureCelsius() float64 {
	return r.Celsius
}

// Label lets TemperatureReading satisfy sensor.Labeled implicitly.
func (r TemperatureReading) Label() string {
	return r.Location
}

// --- Pointer receivers (mutate state in place) ---

// UpdateCelsius changes the Celsius value on the original reading.
func (r *TemperatureReading) UpdateCelsius(celsius float64) {
	r.Celsius = celsius
}

// Relocate changes the location label on the original reading.
func (r *TemperatureReading) Relocate(location string) {
	r.Location = location
}

// Scale multiplies Celsius on the original reading.
func (r *TemperatureReading) Scale(factor float64) {
	r.Celsius *= factor
}

// MethodSetSummary explains which methods are callable on values vs pointers.
func MethodSetSummary() string {
	return `Method sets for TemperatureReading:

Value receivers (callable on T and *T):
  - Fahrenheit()
  - IsFreezing()
  - ScaledCopy(factor)
  - String()

Pointer receivers (only *T implements mutating interfaces):
  - UpdateCelsius(celsius)
  - Relocate(location)
  - Scale(factor)

Rule of thumb:
  - Use value receivers for read-only methods and small structs.
  - Use pointer receivers when mutating state or when any method on the type mutates.
  - For interfaces, pointer-receiver methods mean only *T satisfies the interface.`
}
