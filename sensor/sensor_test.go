package sensor

import (
	"strings"
	"testing"
)

func TestImplicitImplementation(t *testing.T) {
	var _ Measurer = StationReading{Name: "A", Celsius: 1}
	var _ Measurer = ManualGauge{Location: "B", Celsius: 2}
	var _ Labeled = StationReading{}
}

func TestDescribe(t *testing.T) {
	readings := []Measurer{
		StationReading{Name: "Ankara", Celsius: 22.5},
		ManualGauge{Location: "Izmir", Celsius: 18},
	}

	got := PrintReadings(readings)
	if len(got) != 2 || !strings.Contains(got[0], "Ankara") {
		t.Fatalf("unexpected output: %v", got)
	}
}

func TestNilInterfaceValue(t *testing.T) {
	var m Measurer
	if !IsNil(m) {
		t.Fatal("untyped nil interface should be nil")
	}
	if _, ok := SafeMeasure(m); ok {
		t.Fatal("SafeMeasure should fail on nil interface")
	}
}

func TestTypedNilInInterface(t *testing.T) {
	var station *StationReading
	var m Measurer = station

	if m == nil {
		t.Fatal("typed nil pointer in interface is not equal to nil")
	}
	if !IsNil(m) {
		t.Fatal("IsNil should detect typed nil concrete value")
	}
	if _, ok := SafeMeasure(m); ok {
		t.Fatal("SafeMeasure should fail on typed nil pointer")
	}
}

func TestSafeMeasure(t *testing.T) {
	value, ok := SafeMeasure(StationReading{Celsius: 10})
	if !ok || value != 10 {
		t.Fatalf("got (%v, %v), want (10, true)", value, ok)
	}
}
