package calc

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrInvalidOp      = errors.New("invalid operator")
)

func ParseNumber(s string) (float64, error) {
	if s == "" {
		return 0, errors.New("empty input")
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", s, err)
	}
	return n, nil
}

func Calculate(a, b float64, operator string) (float64, error) {
	switch operator {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("%w: %q (use +, -, *, /)", ErrInvalidOp, operator)
	}
}
