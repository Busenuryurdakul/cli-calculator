package main

import (
	"fmt"
	"os"

	"github.com/Busenuryurdakul/cli-calculator/calc"
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

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  cli-calculator <num> <op> <num>       calculator (+, -, *, /)")
	fmt.Fprintln(os.Stderr, "  cli-calculator calc <num> <op> <num> same as above")
	fmt.Fprintln(os.Stderr, "  cli-calculator temp c2f <celsius>     celsius to fahrenheit")
	fmt.Fprintln(os.Stderr, "  cli-calculator temp f2c <fahrenheit>  fahrenheit to celsius")
}
