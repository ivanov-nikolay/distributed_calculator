package calculation

import "errors"

var (
	ErrDivisionByZero  = errors.New("division by zero")
	ErrInvalidOperator = errors.New("invalid operator")
)

var (
	ErrInvalidTask = errors.New("failed to report invalid task")
)

var (
	ErrEmptyExpression    = errors.New("empty expression")
	ErrClosingParenthesis = errors.New("missing closing parenthesis")
)
