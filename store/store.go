package store

import (
	"fmt"
	"sort"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

// Snapshot records length and capacity of a slice.
type Snapshot struct {
	Len int
	Cap int
}

// Stats returns len and cap for a reading slice.
func Stats(items []reading.TemperatureReading) Snapshot {
	return Snapshot{Len: len(items), Cap: cap(items)}
}

// NewLog allocates an empty slice with the given capacity via make.
func NewLog(capacity int) []reading.TemperatureReading {
	return make([]reading.TemperatureReading, 0, capacity)
}

// LiteralLog initializes a slice from a composite literal.
func LiteralLog(items ...reading.TemperatureReading) []reading.TemperatureReading {
	return items
}

// Append adds readings and may grow the backing array when len == cap.
func Append(items []reading.TemperatureReading, extra ...reading.TemperatureReading) []reading.TemperatureReading {
	return append(items, extra...)
}

// Subslice returns items[from:to], a new header over the same backing array.
func Subslice(items []reading.TemperatureReading, from, to int) []reading.TemperatureReading {
	return items[from:to]
}

// IsolatedCopy returns a slice with its own backing array.
func IsolatedCopy(items []reading.TemperatureReading) []reading.TemperatureReading {
	out := make([]reading.TemperatureReading, len(items))
	copy(out, items)
	return out
}

// OverwriteFirst mutates the backing array through a slice header passed by value.
func OverwriteFirst(items []reading.TemperatureReading, celsius float64) {
	if len(items) == 0 {
		return
	}
	items[0].Celsius = celsius
}

// AppendWithinSharedCap appends onto a subslice whose cap still points into
// the original array, so the extra element overwrites the next original slot.
func AppendWithinSharedCap(items []reading.TemperatureReading, extra reading.TemperatureReading) []reading.TemperatureReading {
	return append(items, extra)
}

// NewIndex allocates an empty map with make.
func NewIndex() map[string]reading.TemperatureReading {
	return make(map[string]reading.TemperatureReading)
}

// IndexLiteral initializes a map from a composite literal.
func IndexLiteral(items ...reading.TemperatureReading) map[string]reading.TemperatureReading {
	idx := make(map[string]reading.TemperatureReading, len(items))
	for _, item := range items {
		idx[item.Location] = item
	}
	return idx
}

// Put stores a reading by location. The map header is copied, the backing table is shared.
func Put(idx map[string]reading.TemperatureReading, item reading.TemperatureReading) {
	idx[item.Location] = item
}

// Locations returns map keys from a range loop.
// Map iteration order is unspecified and must not be relied upon.
func Locations(idx map[string]reading.TemperatureReading) []string {
	keys := make([]string, 0, len(idx))
	for location := range idx {
		keys = append(keys, location)
	}
	return keys
}

// LocationsSorted returns map keys in a stable order for tests and display.
func LocationsSorted(idx map[string]reading.TemperatureReading) []string {
	keys := Locations(idx)
	sort.Strings(keys)
	return keys
}

// SumCelsius ranges over a slice (order is the insertion/index order).
func SumCelsius(items []reading.TemperatureReading) float64 {
	var total float64
	for _, item := range items {
		total += item.Celsius
	}
	return total
}

// CollectionSummary documents slice and map behavior in this package.
func CollectionSummary() string {
	return `Slice and map patterns in this package:

- make([]T, 0, n) sets length 0 and capacity n; literals set both from the elements
- append grows length; a new backing array is allocated only when len == cap
- s[i:j] is a new header over the same array; copy() allocates a separate array
- Slice headers and map variables are passed by value but share backing data
- range over a slice is ordered; map iteration order is unspecified and must not be relied upon`
}

// FormatSnapshot is a small helper for CLI output.
func FormatSnapshot(label string, s Snapshot) string {
	return fmt.Sprintf("%s len=%d cap=%d", label, s.Len, s.Cap)
}
