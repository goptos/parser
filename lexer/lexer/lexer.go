package lexer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/goptos/utils"
)

const (
	OpenCodeBracket  string = "0"
	CloseCodeBracket string = "1"
	OpenTagBracket   string = "2"
	CloseTagBracket  string = "3"
	OpenTextBracket  string = "4"
	CloseTextBracket string = "5"
	Number           string = "100"
	Identifier       string = "101"
	Equals           string = "102"
	Colon            string = "103"
	SemiColon        string = "104"
	BinaryOperator   string = "105"
	ForwardSlash     string = "106"
	Space            string = "107"
	Dot              string = "108"
	DoubleQuote      string = "109"
	SingleQuote      string = "110"
	Var              string = "200"
	Func             string = "201"
	On               string = "202"
	If               string = "203"
	EOF              string = "-1"
)

var Keywords = map[string]TokenType{
	// "var":  Var,
	// "func": Func,
	// "on":   On,
	// "if":   If,
}

type TokenType = string

type Token struct {
	V string
	T TokenType
}

func preProcess(source string) string {
	lines := strings.Split(source, "\n")

	var cleanLines = make([]string, 0)
	for i := 0; i < len(lines); i++ {
		cleanLines = append(cleanLines, strings.Trim(lines[i], " "))
	}

	return strings.Join(cleanLines, "\n")
}

func Tokens(source string) (*[]Token, error) {
	utils.Debug("::: base lexer :::\n")
	var tokens = make([]Token, 0)
	var digit = ""
	var word = ""
	for _, r := range preProcess(source) {
		var char = string(r)
		if unicode.IsDigit(r) {
			digit = digit + char
			continue
		}
		if len(digit) > 0 {
			tokens = append(tokens, Token{V: digit, T: Number})
			digit = ""
		}
		if unicode.IsLetter(r) {
			word = word + char
			continue
		}
		if len(word) > 0 {
			reserved, ok := Keywords[word]
			if ok {
				tokens = append(tokens, Token{V: word, T: reserved})
			} else {
				tokens = append(tokens, Token{V: word, T: Identifier})
			}
			word = ""
		}
		switch char {
		case "=":
			tokens = append(tokens, Token{V: char, T: Equals})
		case ":":
			tokens = append(tokens, Token{V: char, T: Colon})
		case ";":
			tokens = append(tokens, Token{V: char, T: SemiColon})
		case "+":
			tokens = append(tokens, Token{V: char, T: BinaryOperator})
		case "-":
			tokens = append(tokens, Token{V: char, T: BinaryOperator})
		case "*":
			tokens = append(tokens, Token{V: char, T: BinaryOperator})
		case "/":
			tokens = append(tokens, Token{V: char, T: ForwardSlash})
		case "%":
			tokens = append(tokens, Token{V: char, T: BinaryOperator})
		case "{":
			tokens = append(tokens, Token{V: char, T: OpenCodeBracket})
		case "}":
			tokens = append(tokens, Token{V: char, T: CloseCodeBracket})
		case "<":
			tokens = append(tokens, Token{V: char, T: OpenTagBracket})
		case ">":
			tokens = append(tokens, Token{V: char, T: CloseTagBracket})
		case "(":
			tokens = append(tokens, Token{V: char, T: OpenTextBracket})
		case ")":
			tokens = append(tokens, Token{V: char, T: CloseTextBracket})
		case "\"":
			tokens = append(tokens, Token{V: char, T: DoubleQuote})
		case ".":
			tokens = append(tokens, Token{V: char, T: Dot})
		case " ":
			tokens = append(tokens, Token{V: char, T: Space})
		case "\t":
		case "\n":
		default:
			return nil, fmt.Errorf("lexer doesn't support '%s' characters", char)
		}
	}
	for i := 0; i < len(tokens); i++ {
		utils.Debug("%d\t%s\t%s\n", i, tokens[i].T, tokens[i].V)
	}
	return &tokens, nil
}
