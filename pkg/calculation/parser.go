package calculation

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	expression string
	position   int
}

type ASTNode struct {
	Operator      string
	Value         float64
	Left, Right   *ASTNode
	TaskScheduled bool
	IsLeaf        bool
}

func (p *parser) parseExpr() (*ASTNode, error) {
	node, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for {
		ch := p.peek()

		if ch == '+' || ch == '-' {
			op := string(p.get())
			right, err := p.parseTerm()

			if err != nil {
				return nil, err
			}

			node = &ASTNode{
				Operator: op,
				Left:     node,
				Right:    right,
				IsLeaf:   false,
			}
		} else {
			break
		}
	}

	return node, nil
}

func (p *parser) peek() rune {
	if p.position < len(p.expression) {
		return rune(p.expression[p.position])
	}

	return 0
}

func (p *parser) get() rune {
	ch := p.peek()
	p.position++

	return ch
}

func (p *parser) parseTerm() (*ASTNode, error) {
	node, err := p.parseFactor()

	if err != nil {
		return nil, err
	}

	for {
		ch := p.peek()

		if ch == '*' || ch == '/' {
			op := string(p.get())
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}

			node = &ASTNode{
				Operator: op,
				Left:     node,
				Right:    right,
				IsLeaf:   false,
			}
		} else {
			break
		}
	}

	return node, nil
}

func (p *parser) parseFactor() (*ASTNode, error) {
	ch := p.peek()

	if ch == '(' {
		p.get()
		node, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, ErrClosingParenthesis
		}
		p.get()
		return node, nil
	}

	start := p.position
	if ch == '+' || ch == '-' {
		p.get()
	}

	for {
		ch = p.peek()
		if unicode.IsDigit(ch) || ch == '.' {
			p.get()
		} else {
			break
		}
	}
	token := p.expression[start:p.position]
	if token == "" {
		return nil, fmt.Errorf("expected number at position: %d\n", start)
	}

	value, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid token's number: %s\n", token)
	}

	return &ASTNode{
		Value:  value,
		IsLeaf: true,
	}, nil
}

func ParseAST(inputExpr string) (*ASTNode, error) {
	expr := strings.ReplaceAll(inputExpr, " ", "")

	if expr == "" {
		return nil, ErrEmptyExpression
	}

	p := &parser{expression: expr, position: 0}

	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	if p.position < len(p.expression) {
		return nil, fmt.Errorf("token unexpected, position: %d\n", p.position)
	}

	return node, nil
}
