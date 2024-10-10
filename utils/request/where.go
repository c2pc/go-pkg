package request

import (
	"fmt"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/clause"
)

func isAndOR(token string) bool {
	// Можно добавить более сложную логику для проверки столбцов, если нужно
	return token == "and" || token == "or"
}

func isOperation(token string) bool {
	operations := []string{"co", "eq", "sw", "ew", "pt", "np", "in", ">", "<", ">=", "<=", "="}
	for _, op := range operations {
		if token == op {
			return true
		}
	}
	return false
}

func parseTokens(tokens []string) (*clause.ExpressionWhere, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	if len(tokens) < 3 {
		return nil, apperr.ErrSyntax.WithError(fmt.Errorf("invalid tokens: %v", tokens))
	}

	if len(tokens) == 3 {
		if !isOperation(tokens[1]) {
			return nil, apperr.ErrSyntax.WithError(fmt.Errorf("invalid operator: %s", tokens[1]))
		}
		return &clause.ExpressionWhere{
			Expressions: nil,
			Column:      tokens[0],
			Operation:   strings.ToLower(tokens[1]),
			Value:       tokens[2],
		}, nil
	}

	var expressions []clause.ExpressionWhere

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "(":
			// Ищем закрывающую скобку
			n := 0
			end := 0
			for j := i + 1; j < len(tokens); j++ {
				if tokens[j] == ")" {
					if n == 0 {
						end = j
						break
					} else {
						n--
					}
				} else if tokens[j] == "(" {
					n++
				}
			}
			// Обработка вложенных выражений
			nestedExpr, err := parseTokens(tokens[i+1 : end])
			if err != nil {
				return nil, err
			}
			if nestedExpr != nil {
				if nestedExpr.Expressions != nil {
					expressions = append(expressions, clause.ExpressionWhere{Expressions: nestedExpr.Expressions})
				} else {
					expressions = append(expressions, *nestedExpr)
				}
			}
			i = end
		default:
			if isOperation(tokens[i]) {
				var value string
				if tokens[i] == "pt" || tokens[i] == "np" {
					value = ""
				} else {
					if i+1 >= len(tokens) {
						return nil, apperr.ErrSyntax.WithError(fmt.Errorf("invalid input operator"))
					}
					value = tokens[i+1]
				}
				expressions = append(expressions, clause.ExpressionWhere{Column: tokens[i-1], Operation: strings.ToLower(tokens[i]), Value: value})
				if tokens[i] != "pt" && tokens[i] != "np" {
					i++
				}

			} else if isAndOR(tokens[i]) {
				expressions = append(expressions, clause.ExpressionWhere{Operation: strings.ToLower(tokens[i])})
			}
		}
	}

	return &clause.ExpressionWhere{Expressions: expressions}, nil
}

func ParseWhere(input string) (*clause.ExpressionWhere, error) {
	if input == "" {
		return nil, nil
	}

	input = addSpaces(input)

	if !validSequence(input) {
		return nil, apperr.ErrSyntax.WithError(fmt.Errorf("invalid input sequence"))
	}

	tokens := makeFields(input)

	return parseTokens(tokens)
}

func addSpaces(input string) string {
	newInput := ""
	isQuote := false

	for _, r := range input {
		if r == '`' {
			if isQuote {
				isQuote = false
				newInput += string(r) + " "
			} else {
				isQuote = true
				newInput += " " + string(r)
			}
			continue
		} else if isQuote {
			newInput += string(r)
			continue
		} else {
			if strings.ContainsRune(",[]()", rune(r)) {
				newInput += " " + string(r) + " "
			} else {
				newInput += string(r)
			}
		}
	}

	return newInput
}

func makeFields(input string) []string {
	var quotes []string
	isQuote := false
	for _, r := range input {
		if !isQuote && r == ' ' {
			continue
		}

		if r == '`' {
			if isQuote {
				isQuote = false
				continue
			} else {
				isQuote = true
				quotes = append(quotes, "")
			}
			continue
		} else if isQuote {
			quotes[len(quotes)-1] += string(r)
			continue
		} else {
			continue
		}
	}

	tokens := strings.Fields(input)

	var newTokens []string
	isQuote = false
	quoteNum := 0
	for _, token := range tokens {
		if token[0] == '`' && token[len(token)-1] == '`' && len(token) != 1 {
			isQuote = false
			newTokens = append(newTokens, "`"+quotes[quoteNum]+"`")
			quoteNum++
		} else if token[0] == '`' || token[len(token)-1] == '`' {
			if isQuote {
				isQuote = false
			} else {
				isQuote = true
				newTokens = append(newTokens, "`"+quotes[quoteNum]+"`")
				quoteNum++
			}
		} else if isQuote {
			continue
		} else {
			if token != "," {
				newTokens = append(newTokens, token)
			}
		}
	}

	var newTokens2 []string
	for i := 0; i < len(newTokens); i++ {
		if newTokens[i] == "[" {
			// Ищем закрывающую скобку
			end := 0
			for j := i + 1; j < len(newTokens); j++ {
				if newTokens[j] == "]" {
					end = j
				}
			}
			newTokens2 = append(newTokens2, strings.Join(newTokens[i+1:end], ","))
			i = end
		} else {
			newTokens2 = append(newTokens2, newTokens[i])
		}
	}

	return newTokens2
}

func validSequence(s string) bool {
	var stack []rune
	bracketMap := map[rune]rune{
		')': '(',
		']': '[',
	}

	newS := ""
	isQuote := false
	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			isQuote = !isQuote
		} else {
			if strings.ContainsRune("()[]", rune(s[i])) && !isQuote {
				newS += string(s[i])
			}
		}
	}

	for _, char := range newS {
		if strings.ContainsRune("([", char) {
			stack = append(stack, char) // открывающая скобка
		} else if strings.ContainsRune(")]", char) {
			if len(stack) == 0 || stack[len(stack)-1] != bracketMap[char] {
				return false // неправильная или несоответствующая скобка
			}
			stack = stack[:len(stack)-1] // закрывающая скобка - удаляем верхнюю
		}
	}

	return len(stack) == 0 // все скобки должны быть закрыты
}
