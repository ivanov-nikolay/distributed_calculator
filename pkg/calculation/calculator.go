package calculation

var (
	SupportedOps = map[string]bool{
		"+": true,
		"-": true,
		"*": true,
		"/": true,
	}
)

func Compute(op string, a, b float64) (float64, error) {
	if !SupportedOps[op] {
		return 0, ErrInvalidOperator
	}

	switch op {
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
		return 0, ErrInvalidOperator
	}
}
