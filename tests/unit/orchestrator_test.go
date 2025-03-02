package unit

import (
	"strings"
	"testing"

	"github.com/ivanov-nikolay/distributed_calculator/internal/service"
)

func normalizeWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func TestTokenizeExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"2 + 3 * 4", []string{"2", "+", "3", "*", "4"}},
		{"(5 + 3) * 4", []string{"(", "5", "+", "3", ")", " ", "*", "4"}},
		{"1 + 2", []string{"1", "+", "2"}},
		{"(1 + 2) * (3 / 4)", []string{"(", "1", "+", "2", ")", " ", "*", " ", "(", "3", "/", "4", ")"}},
		{"2", []string{"2"}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := service.TokenizeExpression(test.input)

			normalizedResult := make([]string, len(result))
			for i, token := range result {
				normalizedResult[i] = normalizeWhitespace(token)
			}

			normalizedExpected := make([]string, len(test.expected))
			for i, token := range test.expected {
				normalizedExpected[i] = normalizeWhitespace(token)
			}

			if len(normalizedResult) != len(normalizedExpected) {
				t.Errorf("expected %v, got %v", normalizedExpected, normalizedResult)
				return
			}

			for i := range normalizedResult {
				if normalizedResult[i] != normalizedExpected[i] {
					t.Errorf("expected %v, got %v", normalizedExpected[i], normalizedResult[i])
				}
			}
		})
	}
}

func TestShuntingYard(t *testing.T) {
	tests := []struct {
		tokens   []string
		expected []string
	}{
		{[]string{"2", "+", "3", "*", "4"}, []string{"2", "3", "4", "*", "+"}},
		{[]string{"(", "5", "+", "3", ")", "*", "4"}, []string{"5", "3", "+", "4", "*"}},
		{[]string{"1", "+", "2"}, []string{"1", "2", "+"}},
		{[]string{"(", "1", "+", "2", ")", "*", "(", "3", "/", "4", ")"}, []string{"1", "2", "+", "3", "4", "/", "*"}},
		{[]string{"2"}, []string{"2"}},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.tokens, " "), func(t *testing.T) {
			result := service.ShuntingYard(test.tokens)
			if !equal(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"2", 2.0},
		{"0", 0.0},
		{"-1", -1.0},
		{"3.14159", 3.14159},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := service.ParseNumber(test.input)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		token    string
		expected bool
	}{
		{"3.14", true},
		{"2", true},
		{"-1", true},
		{"abc", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.token, func(t *testing.T) {
			result := service.IsNumber(test.token)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestIsOperator(t *testing.T) {
	tests := []struct {
		token    string
		expected bool
	}{
		{"+", true},
		{"-", true},
		{"*", true},
		{"/", true},
		{"2", false},
		{"abc", false},
	}

	for _, test := range tests {
		t.Run(test.token, func(t *testing.T) {
			result := service.IsOperator(test.token)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
