package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Busenuryurdakul/cli-calculator/account"
	"github.com/Busenuryurdakul/cli-calculator/calc"
	"github.com/Busenuryurdakul/cli-calculator/logger"
	"github.com/Busenuryurdakul/cli-calculator/pipeline"
	"github.com/Busenuryurdakul/cli-calculator/reading"
	"github.com/Busenuryurdakul/cli-calculator/report"
	"github.com/Busenuryurdakul/cli-calculator/sensor"
	"github.com/Busenuryurdakul/cli-calculator/shape"
	"github.com/Busenuryurdakul/cli-calculator/store"
	"github.com/Busenuryurdakul/cli-calculator/tempconv"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "calc":
		runCalculator(os.Args[2:])
	case "temp":
		runTemperature(os.Args[2:])
	case "demo":
		runStructDemo()
	case "methods":
		runMethodsDemo()
	case "interfaces":
		runInterfacesDemo()
	case "compose":
		runComposeDemo()
	case "library":
		runLibraryDemo()
	case "errors":
		runErrorsDemo()
	case "pipeline":
		if err := runPipeline(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "collections":
		runCollectionsDemo()
	case "files":
		if err := runFilesDemo(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		runCalculator(os.Args[1:])
	}
}

func runCalculator(args []string) {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "calculator: expected 3 arguments: <number> <operator> <number>")
		fmt.Fprintln(os.Stderr, "example: cli-calculator 10 + 5")
		os.Exit(1)
	}

	a, err := calc.ParseNumber(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "calculator: %v\n", err)
		os.Exit(1)
	}

	b, err := calc.ParseNumber(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "calculator: %v\n", err)
		os.Exit(1)
	}

	result, err := calc.Calculate(a, b, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "calculator: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func runTemperature(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "temp: expected 2 arguments: <c2f|f2c> <value>")
		fmt.Fprintln(os.Stderr, "example: cli-calculator temp c2f 100")
		os.Exit(1)
	}

	value, err := calc.ParseNumber(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "temp: %v\n", err)
		os.Exit(1)
	}

	var result float64
	switch args[0] {
	case "c2f":
		result = tempconv.CelsiusToFahrenheit(value)
	case "f2c":
		result = tempconv.FahrenheitToCelsius(value)
	default:
		fmt.Fprintf(os.Stderr, "temp: unknown direction %q (use c2f or f2c)\n", args[0])
		os.Exit(1)
	}

	fmt.Println(result)
}

func runStructDemo() {
	// Positional-style literal (keyed here for clarity; equivalent: {"Izmir", 25.0, "station"})
	literal := reading.DemoPositional()

	// Keyed struct literal
	keyed := reading.TemperatureReading{
		Location: "Ankara",
		Celsius:  22.5,
		Source:   "sensor",
	}

	// Pointer via constructor
	ptr := reading.New("Antalya", 30.0, "api")

	// Pointer to struct literal
	cold := &reading.TemperatureReading{
		Location: "Erzurum",
		Celsius:  -5.0,
		Source:   "manual",
	}

	fmt.Println("=== Default formatting (Go syntax) ===")
	fmt.Printf("%#v\n", literal)

	fmt.Println("\n=== Custom String() output ===")
	fmt.Println(keyed)
	fmt.Println(ptr)

	fmt.Println("\n=== Method preview ===")
	fmt.Printf("%s is freezing: %v\n", cold, cold.IsFreezing())
	cold.UpdateCelsius(2.0)
	fmt.Printf("After UpdateCelsius: %s\n", cold)
}

func runMethodsDemo() {
	original := reading.TemperatureReading{Location: "Ankara", Celsius: 10, Source: "sensor"}

	fmt.Println("=== Value receiver: copy semantics ===")
	scaled := original.ScaledCopy(3)
	fmt.Printf("original: %.1f°C\n", original.Celsius)
	fmt.Printf("scaled copy: %.1f°C\n", scaled.Celsius)

	fmt.Println("\n=== Pointer receiver: mutation ===")
	readingPtr := reading.New("Izmir", 10, "station")
	fmt.Printf("before Scale: %.1f°C\n", readingPtr.Celsius)
	readingPtr.Scale(2)
	fmt.Printf("after Scale: %.1f°C\n", readingPtr.Celsius)

	readingPtr.Relocate("Antalya")
	fmt.Printf("after Relocate: %s\n", readingPtr)

	fmt.Println("\n=== Method sets ===")
	fmt.Println(reading.MethodSetSummary())
}

func runInterfacesDemo() {
	readings := []sensor.Measurer{
		sensor.StationReading{Name: "Ankara", Celsius: 22.5},
		sensor.ManualGauge{Location: "Izmir", Celsius: 18},
		reading.TemperatureReading{Location: "Antalya", Celsius: 30, Source: "api"},
	}

	fmt.Println("=== Interface values (polymorphism) ===")
	for _, line := range sensor.PrintReadings(readings) {
		fmt.Println(line)
	}

	fmt.Println("\n=== Nil interface vs typed nil ===")
	var nilIface sensor.Measurer
	fmt.Printf("untyped nil interface IsNil: %v\n", sensor.IsNil(nilIface))

	var nilStation *sensor.StationReading
	var typedNilIface sensor.Measurer = nilStation
	fmt.Printf("typed nil in interface m == nil: %v\n", typedNilIface == nil)
	fmt.Printf("typed nil detected by IsNil: %v\n", sensor.IsNil(typedNilIface))

	if value, ok := sensor.SafeMeasure(typedNilIface); ok {
		fmt.Println("unexpected value:", value)
	} else {
		fmt.Println("SafeMeasure correctly refused typed nil")
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println(sensor.NilInterfaceSummary())
}

func runComposeDemo() {
	user := account.User{
		Person: account.Person{
			Name:  "Ayse",
			Email: "ayse@example.com",
		},
		LastLogin: "2026-08-15",
	}

	admin := account.Admin{
		User: account.User{
			Person: account.Person{
				Name:  "Mehmet",
				Email: "admin@example.com",
			},
		},
		Department: "Platform",
	}

	fmt.Println("=== Promoted fields ===")
	fmt.Printf("user.Name (from embedded Person): %s\n", user.Name)

	fmt.Println("\n=== Method overrides ===")
	fmt.Printf("Person.Role(): %s\n", user.Person.Role())
	fmt.Printf("User.Role(): %s\n", user.Role())
	fmt.Printf("Admin.Role(): %s\n", admin.Role())

	fmt.Println("\n=== Composed domain via interface ===")
	accounts := []account.Account{user, admin}
	for _, a := range accounts {
		fmt.Printf("%s -> role=%s dashboard=%v settings=%v\n",
			a.Describe(),
			a.Role(),
			a.CanAccess("dashboard"),
			a.CanAccess("settings"),
		)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println(account.ComposeSummary())
}

func runLibraryDemo() {
	shapes := []shape.Shape{
		shape.Circle{Radius: 3},
		shape.Rectangle{Width: 4, Height: 5},
	}

	fmt.Println("=== Shape library ===")
	for _, s := range shapes {
		fmt.Printf("area=%.2f perimeter=%.2f\n", s.Area(), s.Perimeter())
	}

	fmt.Println("\n=== Logger abstraction ===")
	loggers := []logger.Logger{
		logger.ConsoleLogger{},
		logger.NoopLogger{},
	}
	for _, log := range loggers {
		logger.LogStartup(log, "cli-calculator")
	}

	fmt.Println("\n=== Code review notes ===")
	fmt.Println(shape.ReviewSummary())
	fmt.Println(logger.ReviewSummary())
}

func runErrorsDemo() {
	examples := []struct {
		title                           string
		location, celsius, source, unit string
	}{
		{title: "parse failure", location: "Ankara", celsius: "abc", source: "sensor", unit: "C"},
		{title: "empty location", location: "", celsius: "22.5", source: "sensor", unit: "C"},
		{title: "below absolute zero", location: "Space", celsius: "-400", source: "probe", unit: "C"},
		{title: "invalid unit", location: "Izmir", celsius: "18", source: "manual", unit: "K"},
		{title: "success", location: "Ankara", celsius: "22.5", source: "sensor", unit: "F"},
	}

	fmt.Println("=== Return (T, error) and wrap with %w ===")
	for _, ex := range examples {
		result, err := pipeline.Process(ex.location, ex.celsius, ex.source, ex.unit)
		if err != nil {
			fmt.Printf("%s: %v\n", ex.title, err)
			fmt.Printf("  classify: %s\n", pipeline.Classify(err))
			fmt.Printf("  errors.Is BelowAbsoluteZero: %v\n", errors.Is(err, pipeline.ErrBelowAbsoluteZero))

			var parseErr *pipeline.ParseError
			if errors.As(err, &parseErr) {
				fmt.Printf("  errors.As ParseError: input=%q inner=%v\n", parseErr.Input, parseErr.Err)
			}
			continue
		}
		fmt.Printf("%s: %s -> %.1f°%s\n", ex.title, result.Reading, result.Converted, strings.ToUpper(result.Unit[:1]))
	}

	fmt.Println("\n=== Fail fast ===")
	fmt.Println("Process returns on the first error (parse, then validate, then convert).")
	fmt.Println("That keeps handlers flat: if err != nil { return ..., err }")

	fmt.Println("\n=== Summary ===")
	fmt.Println(pipeline.ErrorSummary())
}

func runPipeline(args []string) error {
	if len(args) < 3 || len(args) > 4 {
		return errors.New("pipeline: expected <location> <celsius> <source> [C|F]")
	}

	unit := "C"
	if len(args) == 4 {
		unit = args[3]
	}

	result, err := pipeline.Process(args[0], args[1], args[2], unit)
	if err != nil {
		return fmt.Errorf("pipeline (%s): %w", pipeline.Classify(err), err)
	}

	fmt.Printf("%s -> %.1f°%s\n", result.Reading, result.Converted, strings.ToUpper(result.Unit[:1]))
	return nil
}

func runCollectionsDemo() {
	made := store.NewLog(4)
	fmt.Println("=== make vs literal ===")
	fmt.Println(store.FormatSnapshot("make([]T, 0, 4)", store.Stats(made)))

	lit := store.LiteralLog(
		reading.TemperatureReading{Location: "Ankara", Celsius: 10, Source: "sensor"},
		reading.TemperatureReading{Location: "Izmir", Celsius: 20, Source: "station"},
		reading.TemperatureReading{Location: "Antalya", Celsius: 30, Source: "api"},
	)
	fmt.Println(store.FormatSnapshot("literal 3 readings", store.Stats(lit)))

	grown := store.Append(made,
		reading.TemperatureReading{Location: "Ankara", Celsius: 10, Source: "sensor"},
		reading.TemperatureReading{Location: "Izmir", Celsius: 20, Source: "station"},
	)
	fmt.Println("\n=== append, slice, copy ===")
	fmt.Println(store.FormatSnapshot("append 2 into cap 4", store.Stats(grown)))
	grown = store.Append(grown,
		reading.TemperatureReading{Location: "Antalya", Celsius: 30, Source: "api"},
		reading.TemperatureReading{Location: "Bursa", Celsius: 12, Source: "manual"},
		reading.TemperatureReading{Location: "Van", Celsius: -2, Source: "station"},
	)
	fmt.Println(store.FormatSnapshot("append past capacity", store.Stats(grown)))

	view := store.Subslice(lit, 0, 2)
	cloned := store.IsolatedCopy(lit)
	fmt.Println(store.FormatSnapshot("subslice [0:2]", store.Stats(view)))
	fmt.Println(store.FormatSnapshot("isolated copy", store.Stats(cloned)))

	fmt.Println("\n=== shared mutation surprises ===")
	fmt.Printf("before overwrite: original[0]=%.0f copy[0]=%.0f\n", lit[0].Celsius, cloned[0].Celsius)
	store.OverwriteFirst(view, 99)
	fmt.Printf("after overwrite via subslice: original[0]=%.0f copy[0]=%.0f\n", lit[0].Celsius, cloned[0].Celsius)

	shared := store.LiteralLog(
		reading.TemperatureReading{Location: "A", Celsius: 1, Source: "demo"},
		reading.TemperatureReading{Location: "B", Celsius: 2, Source: "demo"},
		reading.TemperatureReading{Location: "C", Celsius: 3, Source: "demo"},
	)
	head := store.Subslice(shared, 0, 2)
	_ = store.AppendWithinSharedCap(head, reading.TemperatureReading{Location: "X", Celsius: 99, Source: "demo"})
	fmt.Printf("append on subslice overwrote shared[2]: %+v\n", shared[2])

	fmt.Println("\n=== range over slice (stable order) ===")
	for i, item := range lit {
		fmt.Printf("%d: %s %.0f°C\n", i, item.Location, item.Celsius)
	}

	idx := store.IndexLiteral(lit...)
	store.Put(idx, reading.TemperatureReading{Location: "Erzurum", Celsius: -5, Source: "manual"})
	fmt.Println("\n=== range over map (order is unspecified) ===")
	fmt.Printf("range keys (do not rely on order): %v\n", store.Locations(idx))
	fmt.Printf("sorted keys:                       %v\n", store.LocationsSorted(idx))

	fmt.Println("\n=== Summary ===")
	fmt.Println(store.CollectionSummary())
}

func runFilesDemo() error {
	dir, err := os.MkdirTemp("", "cli-calculator-report-*")
	if err != nil {
		return fmt.Errorf("files demo: %w", err)
	}
	defer os.RemoveAll(dir)

	items := []reading.TemperatureReading{
		{Location: "Ankara", Celsius: 10, Source: "sensor"},
		{Location: "Izmir", Celsius: 20, Source: "station"},
		{Location: "Antalya", Celsius: 30, Source: "api"},
	}

	allPath := filepath.Join(dir, "all.txt")
	linesPath := filepath.Join(dir, "lines.txt")
	copyPath := filepath.Join(dir, "copy.txt")

	fmt.Println("=== Write files ===")
	if err := report.WriteAll(allPath, items); err != nil {
		return err
	}
	if err := report.WriteBuffered(linesPath, items); err != nil {
		return err
	}
	fmt.Println("os.WriteFile:", allPath)
	fmt.Println("bufio.Writer:", linesPath)

	fmt.Println("\n=== Read files ===")
	fromAll, err := report.ReadAll(allPath)
	if err != nil {
		return err
	}
	fromLines, err := report.ReadLines(linesPath)
	if err != nil {
		return err
	}
	fmt.Printf("os.ReadFile: %d readings\n", len(fromAll))
	fmt.Printf("bufio.Scanner: %d readings\n", len(fromLines))
	for i, item := range fromLines {
		fmt.Printf("%d: %s\n", i, item)
	}

	fmt.Println("\n=== io.Reader / io.Writer ===")
	src, err := os.Open(allPath)
	if err != nil {
		return fmt.Errorf("files demo: %w", err)
	}
	dst, err := os.Create(copyPath)
	if err != nil {
		src.Close()
		return fmt.Errorf("files demo: %w", err)
	}
	n, err := report.Copy(dst, src)
	src.Close()
	closeErr := dst.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return fmt.Errorf("files demo: %w", closeErr)
	}
	fmt.Printf("io.Copy wrote %d bytes to %s\n", n, copyPath)

	copied, err := os.Open(copyPath)
	if err != nil {
		return fmt.Errorf("files demo: %w", err)
	}
	defer copied.Close()
	lines, err := report.CountLines(copied)
	if err != nil {
		return err
	}
	fmt.Printf("bufio line count: %d\n", lines)

	fmt.Println("\n=== Missing file error ===")
	_, err = report.ReadAll(filepath.Join(dir, "missing.txt"))
	fmt.Println(err)
	fmt.Printf("errors.Is(err, os.ErrNotExist): %v\n", errors.Is(err, os.ErrNotExist))

	fmt.Println("\n=== Summary ===")
	fmt.Println(report.FileSummary())
	return nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  cli-calculator <num> <op> <num>       calculator (+, -, *, /)")
	fmt.Fprintln(os.Stderr, "  cli-calculator calc <num> <op> <num> same as above")
	fmt.Fprintln(os.Stderr, "  cli-calculator temp c2f <celsius>     celsius to fahrenheit")
	fmt.Fprintln(os.Stderr, "  cli-calculator temp f2c <fahrenheit>  fahrenheit to celsius")
	fmt.Fprintln(os.Stderr, "  cli-calculator demo                   struct initialization demo")
	fmt.Fprintln(os.Stderr, "  cli-calculator methods                value vs pointer receiver demo")
	fmt.Fprintln(os.Stderr, "  cli-calculator interfaces             small interfaces and nil demo")
	fmt.Fprintln(os.Stderr, "  cli-calculator compose                struct embedding and composition demo")
	fmt.Fprintln(os.Stderr, "  cli-calculator library                shape and logger library demo")
	fmt.Fprintln(os.Stderr, "  cli-calculator errors                 error values, wrapping, Is/As demo")
	fmt.Fprintln(os.Stderr, "  cli-calculator pipeline <loc> <c> <src> [C|F]  fail-fast reading pipeline")
	fmt.Fprintln(os.Stderr, "  cli-calculator collections            slices, maps, make, shared mutation")
	fmt.Fprintln(os.Stderr, "  cli-calculator files                  read/write reports with io and bufio")
}
