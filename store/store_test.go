package store

import (
	"reflect"
	"testing"

	"github.com/Busenuryurdakul/cli-calculator/reading"
)

func sample(location string, celsius float64) reading.TemperatureReading {
	return reading.TemperatureReading{Location: location, Celsius: celsius, Source: "station"}
}

func TestMakeVsLiteral(t *testing.T) {
	made := NewLog(4)
	if got := Stats(made); got != (Snapshot{Len: 0, Cap: 4}) {
		t.Fatalf("make: got %+v, want len=0 cap=4", got)
	}

	lit := LiteralLog(sample("Ankara", 10), sample("Izmir", 20))
	if got := Stats(lit); got != (Snapshot{Len: 2, Cap: 2}) {
		t.Fatalf("literal: got %+v, want len=2 cap=2", got)
	}
}

func TestAppendGrowsCapacityAndKeepsElements(t *testing.T) {
	items := NewLog(2)
	items = Append(items, sample("Ankara", 10))
	if got := Stats(items); got != (Snapshot{Len: 1, Cap: 2}) {
		t.Fatalf("within cap: got %+v", got)
	}

	oldCap := cap(items)
	items = Append(items, sample("Izmir", 20), sample("Antalya", 30))
	if len(items) != 3 {
		t.Fatalf("after grow: len=%d, want 3", len(items))
	}
	if cap(items) <= oldCap {
		t.Fatalf("after grow: cap=%d, want greater than %d", cap(items), oldCap)
	}
	if items[0].Location != "Ankara" || items[1].Location != "Izmir" || items[2].Location != "Antalya" {
		t.Fatalf("elements not preserved after grow: %+v", items)
	}
}

func TestSubsliceSharesBackingArray(t *testing.T) {
	orig := LiteralLog(sample("Ankara", 10), sample("Izmir", 20), sample("Antalya", 30))
	view := Subslice(orig, 0, 2)

	OverwriteFirst(view, 99)

	if orig[0].Celsius != 99 {
		t.Fatalf("original should change through shared backing, got %v", orig[0].Celsius)
	}
}

func TestIsolatedCopyDoesNotShare(t *testing.T) {
	orig := LiteralLog(sample("Ankara", 10), sample("Izmir", 20))
	cloned := IsolatedCopy(orig)

	OverwriteFirst(cloned, 99)

	if orig[0].Celsius != 10 {
		t.Fatalf("copy should not mutate original, got %v", orig[0].Celsius)
	}
	if cloned[0].Celsius != 99 {
		t.Fatalf("copy should be mutated, got %v", cloned[0].Celsius)
	}
}

func TestAppendWithinSharedCapOverwrites(t *testing.T) {
	orig := LiteralLog(sample("Ankara", 10), sample("Izmir", 20), sample("Antalya", 30))
	view := Subslice(orig, 0, 2) // len=2, cap>=3, same array

	_ = AppendWithinSharedCap(view, sample("Bursa", 99))

	if orig[2].Location != "Bursa" || orig[2].Celsius != 99 {
		t.Fatalf("append within shared cap should overwrite orig[2], got %+v", orig[2])
	}
}

func TestRangeSliceIsOrdered(t *testing.T) {
	items := LiteralLog(sample("Ankara", 1), sample("Izmir", 2), sample("Antalya", 3))
	if SumCelsius(items) != 6 {
		t.Fatalf("sum got %v, want 6", SumCelsius(items))
	}

	got := make([]string, 0, len(items))
	for _, item := range items {
		got = append(got, item.Location)
	}
	want := []string{"Ankara", "Izmir", "Antalya"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("slice range order got %v, want %v", got, want)
	}
}

func TestMapLiteralAndMake(t *testing.T) {
	empty := NewIndex()
	if len(empty) != 0 {
		t.Fatalf("make map should start empty, got %d", len(empty))
	}

	idx := IndexLiteral(sample("Ankara", 10), sample("Izmir", 20))
	if idx["Ankara"].Celsius != 10 || idx["Izmir"].Celsius != 20 {
		t.Fatalf("unexpected index: %+v", idx)
	}
}

func TestMapSharedMutation(t *testing.T) {
	idx := IndexLiteral(sample("Ankara", 10))
	Put(idx, sample("Ankara", 42))

	if idx["Ankara"].Celsius != 42 {
		t.Fatalf("map backing table is shared; got %v", idx["Ankara"].Celsius)
	}
}

func TestMapKeysStableWhenSorted(t *testing.T) {
	idx := IndexLiteral(sample("Ankara", 10), sample("Izmir", 20), sample("Antalya", 30))

	// Do not assert range order: map iteration order is unspecified
	// and must not be relied upon. Use LocationsSorted for stable output.
	got := LocationsSorted(idx)
	want := []string{"Ankara", "Antalya", "Izmir"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted keys got %v, want %v", got, want)
	}
}
