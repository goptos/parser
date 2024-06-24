package lexer

import (
	"fmt"
	"strings"

	"github.com/goptos/parser/lexer/lexer"
	"github.com/goptos/utils"
)

const (
	Code      string = "Code"
	Quote     string = "Quote"
	CloseTag  string = "TagC"
	OpenTag   string = "TagO"
	Comp      string = "Comp"
	EventAttr string = "AttrE"
	Attr      string = "Attr"
	Arg       string = "Arg"
	Text      string = "Text"
)

type TokenType = lexer.TokenType
type Token = lexer.Token

type Increment struct {
	j int
	u int
}

func (_self *Increment) New(ubound int) *Increment {
	var inc = new(Increment)
	inc.u = ubound - 1
	return inc
}

func (_self *Increment) Peak() int {
	var j = _self.j
	return j
}

func (_self *Increment) Get(i int) int {
	var j = _self.j
	if i+j > _self.u {
		return _self.u
	}
	if j+1 <= _self.u {
		_self.j += 1
	}
	return i + j
}

func (_self *Increment) Reset() {
	_self.j = 0
}

func collect(t []Token, i int, j int) string {
	var hit = ""
	var picked = utils.Pick(t, i, j)
	for k := 0; k < len(picked); k++ {
		hit = hit + picked[k].V
	}
	return hit
}

func printLastToken(tokens []Token) {
	utils.Debug("%d\t%s\t%s\n", len(tokens)-1, tokens[len(tokens)-1].T, tokens[len(tokens)-1].V)
}

func collectCodeBlocks(t []Token) []Token {
	var tokens = []Token{}
	var j = (*Increment).New(nil, len(t))
	for i := 0; i < len(t); {
		var hit = false
		if t[j.Get(i)].T == lexer.OpenCodeBracket {
			hit = true
			var depth = 1
			for k := i; k < len(t); k++ {
				if t[i+j.Peak()].T == lexer.OpenCodeBracket {
					depth++
				}
				if t[j.Get(i)].T == lexer.CloseCodeBracket {
					depth--
				}
				if depth == 0 {
					break
				}
			}
			tokens = append(tokens, Token{
				V: collect(t, i, i+j.Peak()-1),
				T: Code})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()
		if !hit {
			tokens = append(tokens, t[i])
			i++
		}
	}
	return tokens
}

func collectDoubleQuoteSections(t []Token) []Token {
	var tokens = []Token{}
	var j = (*Increment).New(nil, len(t))
	for i := 0; i < len(t); {
		var hit = false
		if t[j.Get(i)].T == lexer.DoubleQuote {
			hit = true
			var depth = 1
			for k := i; k < len(t); k++ {
				if t[j.Get(i)].T == lexer.DoubleQuote {
					depth--
				}
				if depth == 0 {
					break
				}
			}
			tokens = append(tokens, Token{
				V: collect(t, i, i+j.Peak()-1),
				T: Quote})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()
		if !hit {
			tokens = append(tokens, t[i])
			i++
		}
	}
	return tokens
}

func Tokens(source string) (*[]Token, error) {
	var baseTokens, err = lexer.Tokens(source)
	if err != nil {
		return nil, err
	}

	utils.Debug("::: lexer :::\n")
	var t = *baseTokens

	utils.Debug(" :: code ::\n")
	t = collectCodeBlocks(t)

	utils.Debug(" :: quotes ::\n")
	t = collectDoubleQuoteSections(t)

	if utils.DEBUG {
		utils.Debug(" :: new base ::\n")
		for i := 0; i < len(t); i++ {
			utils.Debug("%d\t%s\t%s\n", i, t[i].T, t[i].V)
		}
	}

	utils.Debug(" :: tags ::\n")
	var tokens = make([]Token, 0)
	var j = (*Increment).New(nil, len(t))
	for i := 0; i < len(t); {
		var hit = false

		// Find closing html tags `</div>`
		if !hit &&
			t[j.Get(i)].T == lexer.OpenTagBracket &&
			t[j.Get(i)].T == lexer.ForwardSlash &&
			t[j.Get(i)].T == lexer.Identifier &&
			t[j.Get(i)].T == lexer.CloseTagBracket {
			hit = true
			tokens = append(tokens, Token{
				V: collect(t, i+2, i+2),
				T: CloseTag})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()

		// Find component tags `<button />`
		if !hit &&
			t[j.Get(i)].T == lexer.OpenTagBracket &&
			t[j.Get(i)].T == lexer.Identifier &&
			t[j.Get(i)].T == lexer.Space &&
			t[j.Get(i)].T == lexer.ForwardSlash &&
			t[j.Get(i)].T == lexer.CloseTagBracket {
			hit = true
			tokens = append(tokens, Token{
				V: collect(t, i+1, i+1),
				T: Comp})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()

		// Find opening html tags `<div>`
		if !hit &&
			t[j.Get(i)].T == lexer.OpenTagBracket &&
			t[j.Get(i)].T == lexer.Identifier &&
			t[j.Get(i)].T == lexer.CloseTagBracket {
			hit = true
			tokens = append(tokens, Token{
				V: collect(t, i+1, i+1),
				T: OpenTag})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()

		// Find opening html tags with attr `<button` and component tags with args `<Button`
		if !hit &&
			t[j.Get(i)].T == lexer.OpenTagBracket &&
			t[j.Get(i)].T == lexer.Identifier {
			hit = true
			if string(t[i+1].V[0]) == strings.ToUpper(string(t[i+1].V[0])) {
				tokens = append(tokens, Token{
					V: collect(t, i+1, i+1),
					T: Comp})
			} else {
				tokens = append(tokens, Token{
					V: collect(t, i+1, i+1),
					T: OpenTag})
			}
			printLastToken(tokens)
			i += j.Peak()
			j.Reset()

			for {
				var innerHit = false

				// Find event attr ` on:click=`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Space &&
					t[j.Get(i)].T == lexer.Identifier &&
					t[j.Get(i)].T == lexer.Colon &&
					t[j.Get(i)].T == lexer.Identifier &&
					t[j.Get(i)].T == lexer.Equals {
					innerHit = true
					tokens = append(tokens, Token{
						V: collect(t, i+1, i+3),
						T: EventAttr})
					printLastToken(tokens)
					i += j.Peak() - 1
				}
				j.Reset()

				// Find attr ` href=`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Space &&
					t[j.Get(i)].T == lexer.Identifier &&
					t[j.Get(i)].T == lexer.Equals {
					innerHit = true
					tokens = append(tokens, Token{
						V: collect(t, i+1, i+1),
						T: Attr})
					printLastToken(tokens)
					i += j.Peak() - 1
				}
				j.Reset()

				// Find attr code value `={...}`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Equals &&
					t[j.Get(i)].T == Code {
					innerHit = true
					tokens = append(tokens, Token{
						V: collect(t, i+1, i+1),
						T: Code})
					printLastToken(tokens)
					i += j.Peak()
				}
				j.Reset()

				// Find attr quote value `="..."`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Equals &&
					t[j.Get(i)].T == Quote {
					innerHit = true
					tokens = append(tokens, Token{
						V: collect(t, i+1, i+1),
						T: Quote})
					printLastToken(tokens)
					i += j.Peak()
				}
				j.Reset()

				// Find arg ` arg`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Space &&
					t[j.Get(i)].T == lexer.Identifier {
					innerHit = true
					tokens = append(tokens, Token{
						V: collect(t, i+1, i+1),
						T: Arg})
					printLastToken(tokens)
					i += j.Peak()
				}
				j.Reset()

				// find component tag bracket ` />`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Space &&
					t[j.Get(i)].T == lexer.ForwardSlash &&
					t[j.Get(i)].T == lexer.CloseTagBracket {
					innerHit = true
					i += j.Peak()
					break
				}
				j.Reset()

				// find component tag bracket `/>`
				if !innerHit &&
					t[j.Get(i)].T == lexer.ForwardSlash &&
					t[j.Get(i)].T == lexer.CloseTagBracket {
					innerHit = true
					i += j.Peak()
					break
				}
				j.Reset()

				// find html tag bracket ` >`
				if !innerHit &&
					t[j.Get(i)].T == lexer.Space &&
					t[j.Get(i)].T == lexer.CloseTagBracket {
					innerHit = true
					i += j.Peak()
					break
				}
				j.Reset()

				// find html tag bracket `>`
				if !innerHit &&
					t[j.Get(i)].T == lexer.CloseTagBracket {
					innerHit = true
					i += j.Peak()
					break
				}
				j.Reset()

				// find space ` `
				if !innerHit &&
					t[j.Get(i)].T == lexer.Space {
					innerHit = true
					i += j.Peak()
				}
				j.Reset()

				if !innerHit {
					i++
					utils.Debug("unknown inner token at '%d' type '%s' value '%s'\n", i-1, t[i-1].T, t[i-1].V)
					return nil, fmt.Errorf("unknown inner token at '%d' type '%s' value '%s'", i-1, t[i-1].T, t[i-1].V)
				}
			}
		}
		j.Reset()

		if !hit &&
			t[j.Get(i)].T == Code {
			hit = true
			tokens = append(tokens, Token{
				V: collect(t, i, i),
				T: Code})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()

		if !hit &&
			t[j.Get(i)].T == lexer.Identifier {
			hit = true
			tokens = append(tokens, Token{
				V: collect(t, i, i),
				T: Text})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()

		if !hit &&
			t[j.Get(i)].T == lexer.Space {
			hit = true
			tokens = append(tokens, Token{
				V: collect(t, i, i),
				T: Text})
			printLastToken(tokens)
			i += j.Peak()
		}
		j.Reset()

		if !hit {
			i++
			utils.Debug("unknown token at '%d' type '%s' value '%s'\n", i-1, t[i-1].T, t[i-1].V)
			return nil, fmt.Errorf("unknown token at '%d' type '%s' value '%s'", i-1, t[i-1].T, t[i-1].V)
		}
	}

	return &tokens, nil
}
