package service

import (
	"strconv"
	"strings"
)

// TokenizeExpression разделяет выражение на токены
func TokenizeExpression(expr string) []string {
	var tokens []string
	var buffer strings.Builder

	for _, char := range expr {
		if IsOperator(string(char)) || char == '(' || char == ')' {
			if buffer.Len() > 0 {
				tokens = append(tokens, buffer.String())
				buffer.Reset()
			}
			tokens = append(tokens, string(char))
		} else {
			buffer.WriteRune(char)
		}
	}

	if buffer.Len() > 0 {
		tokens = append(tokens, buffer.String())
	}

	return tokens
}

// ShuntingYard преоброзовывает набор токенов в RPN
func ShuntingYard(tokens []string) []string {
	var output []string
	var operators []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
		"(": 0, // Наименьший приоритет для открывающей скобки
	}

	for _, token := range tokens {
		if IsNumber(token) {
			output = append(output, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			// Выталкиваем все операторы до открывающей скобки
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			// Удаляем открывающую скобку из стека
			if len(operators) > 0 && operators[len(operators)-1] == "(" {
				operators = operators[:len(operators)-1]
			}
		} else if IsOperator(token) {
			for len(operators) > 0 && precedence[operators[len(operators)-1]] >= precedence[token] {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		}
	}

	// Выталкиваем оставшиеся операторы из стека
	for len(operators) > 0 {
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output
}

// ParseNumber преобразовывает строку (число) в вещественное число
func ParseNumber(s string) float64 {
	num, _ := strconv.ParseFloat(s, 64)
	return num
}

// IsNumber проверяет токен на соответствие вещесственному числу
func IsNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// IsOperator проверяет токен на соответствие математическому оператору
func IsOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}
