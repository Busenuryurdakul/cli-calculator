package reading

import "testing"

func TestValueReceiverCopySemantics(t *testing.T) {
	original := TemperatureReading{Location: "Ankara", Celsius: 10, Source: "sensor"}
	scaled := original.ScaledCopy(3)

	if original.Celsius != 10 {
		t.Fatalf("original was mutated: got %v, want 10", original.Celsius)
	}
	if scaled.Celsius != 30 {
		t.Fatalf("scaled copy wrong: got %v, want 30", scaled.Celsius)
	}
}

func TestPointerReceiverMutates(t *testing.T) {
	r := New("Izmir", 10, "station")
	r.Scale(2)

	if r.Celsius != 20 {
		t.Fatalf("pointer receiver did not mutate: got %v, want 20", r.Celsius)
	}
}

func TestRelocate(t *testing.T) {
	r := New("Izmir", 15, "manual")
	r.Relocate("Bursa")

	if r.Location != "Bursa" {
		t.Fatalf("got location %q, want Bursa", r.Location)
	}
}

func TestMethodSetValueReceiverInterface(t *testing.T) {
	var valueChecker FreezingChecker = TemperatureReading{Celsius: -1}
	if !valueChecker.IsFreezing() {
		t.Fatal("value should satisfy FreezingChecker")
	}

	var pointerChecker FreezingChecker = &TemperatureReading{Celsius: -1}
	if !pointerChecker.IsFreezing() {
		t.Fatal("pointer should satisfy FreezingChecker")
	}
}

func TestMethodSetPointerReceiverInterface(t *testing.T) {
	var updater CelsiusUpdater = &TemperatureReading{Celsius: 5}
	updater.UpdateCelsius(12)

	if updater.(*TemperatureReading).Celsius != 12 {
		t.Fatalf("got %v, want 12", updater.(*TemperatureReading).Celsius)
	}
}

func TestMethodSetSummary(t *testing.T) {
	if MethodSetSummary() == "" {
		t.Fatal("expected non-empty method set summary")
	}
}
