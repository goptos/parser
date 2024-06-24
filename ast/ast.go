package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goptos/parser/lexer"
	"github.com/goptos/utils"
)

type TokenType = lexer.TokenType
type Token = lexer.Token
type Increment = lexer.Increment

const (
	xHtmlElem     int = 0
	HtmlEventAttr int = 1
	xHtmlAttr     int = 2
	HtmlText      int = 3
	HtmlDynText   int = 100
)

type NodeType interface {
	ElemNode |
		CompNode |
		EndNode |
		TextNode |
		DynTextNode |
		AttrNode |
		EventAttrNode
}

type Node[T NodeType] struct {
	Type  TokenType
	Value T
}

type ElemNode struct {
	Name     string
	Children []interface{} // []Node
}
type CompNode struct {
	Name string
	Args []string
}
type EndNode struct {
	Name string
	Elem ElemNode
}
type TextNode struct {
	Data string
}
type DynTextNode struct {
	Effect string
}
type AttrNode struct {
	Name      string
	Value     string
	ValueType string
}
type EventAttrNode struct {
	Name       string
	Event      string
	Effect     string
	EffectType string
}
type Ast struct {
	Root                   Node[ElemNode]
	ElemNodeProcessor      func(ElemNode)
	CompNodeNodeProcessor  func(CompNode)
	EndNodeProcessor       func(EndNode)
	TextNodeProcessor      func(TextNode)
	DynTextNodeProcessor   func(DynTextNode)
	AttrNodeProcessor      func(AttrNode)
	EventAttrNodeProcessor func(EventAttrNode)
}

func (_self *Ast) New(source string) (*Ast, error) {
	tokens, err := lexer.Tokens(source)
	if err != nil {
		return nil, err
	}
	utils.Debug("::: ast create :::\n")
	var t = *tokens
	if t[0].T != lexer.OpenTag {
		return nil, errors.New("not valid HTML file")
	}
	var index = 0
	root, err := createNode(&t, &index)
	if err != nil {
		return nil, err
	}

	return &Ast{
		Root:                   Node[ElemNode]{Type: lexer.OpenTag, Value: *root},
		ElemNodeProcessor:      printElemNode,
		CompNodeNodeProcessor:  printCompNode,
		EndNodeProcessor:       printEndNode,
		TextNodeProcessor:      printTextNode,
		DynTextNodeProcessor:   printDynTextNode,
		AttrNodeProcessor:      printAttrNode,
		EventAttrNodeProcessor: printEventAttrNode,
	}, nil
}

func createNode(tokens *[]Token, index *int) (*ElemNode, error) {
	var t = *tokens
	utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
	var el = ElemNode{Name: t[*index].V, Children: make([]interface{}, 0)}
	*index++

	for *index < len(t) {
		var n = t[*index]
		if n.T == lexer.OpenTag {
			var child, err = createNode(&t, index)
			if err != nil {
				return nil, err
			}
			el.Children = append(el.Children,
				Node[ElemNode]{
					Type:  n.T,
					Value: *child})
			*index++
		}
		if n.T == lexer.Comp {
			utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			var args = []string{}
			for *index+1 < len(t) {
				if t[*index+1].T != lexer.Arg {
					break
				}
				*index++
				args = append(args, t[*index].V)
				utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			}
			el.Children = append(el.Children,
				Node[CompNode]{
					Type: n.T,
					Value: CompNode{
						Name: n.V,
						Args: args}})
			*index++
		}
		if n.T == lexer.Attr {
			utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			for *index+1 < len(t) {
				if t[*index+1].T == lexer.Quote || t[*index+1].T == lexer.Code {
					*index++
					break
				}
				return nil, fmt.Errorf("attr '%s' missing quote value (index %d)", n.V, *index)
			}
			var next = t[*index]
			utils.Debug("%d\t%s\t%s\n", *index, next.T, next.V)
			el.Children = append(el.Children,
				Node[AttrNode]{
					Type: n.T,
					Value: AttrNode{
						Name:      n.V,
						Value:     next.V,
						ValueType: next.T}})
			*index++
		}
		if n.T == lexer.EventAttr {
			utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			for *index+1 < len(t) {
				if t[*index+1].T == lexer.Code {
					*index++
					break
				}
				return nil, fmt.Errorf("event attr '%s' missing code value (index %d)", n.V, *index)
			}
			var next = t[*index]
			utils.Debug("%d\t%s\t%s\n", *index, next.T, next.V)
			el.Children = append(el.Children,
				Node[EventAttrNode]{
					Type: n.T,
					Value: EventAttrNode{
						Name:       strings.Split(n.V, ":")[0],
						Event:      strings.Split(n.V, ":")[1],
						Effect:     next.V,
						EffectType: next.T}})
			*index++
		}
		if n.T == lexer.Text {
			utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			el.Children = append(el.Children,
				Node[TextNode]{
					Type: n.T,
					Value: TextNode{
						Data: n.V}})
			*index++
		}
		if n.T == lexer.Code {
			utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			el.Children = append(el.Children,
				Node[DynTextNode]{
					Type: n.T,
					Value: DynTextNode{
						Effect: n.V}})
			*index++
		}
		if n.T == lexer.CloseTag {
			utils.Debug("%d\t%s\t%s\n", *index, t[*index].T, t[*index].V)
			el.Children = append(el.Children,
				Node[EndNode]{
					Type: n.T,
					Value: EndNode{
						Name: n.V,
						Elem: el}})
			return &el, nil
		}
	}

	return nil, fmt.Errorf("no closing tag for elem `%s`", el.Name)
}

func (_self *Ast) Process() {
	utils.Debug("::: ast process :::\n")
	_self.processR(_self.Root)
}

func printElemNode(node ElemNode) {
	utils.Debug("%s\t%s\n",
		lexer.OpenTag,
		node.Name)
}

func printCompNode(node CompNode) {
	utils.Debug("%s\t%s\t%s\n",
		lexer.OpenTag,
		node.Name,
		strings.Join(node.Args, "\t"))
}

func printEndNode(node EndNode) {
	utils.Debug("%s\t%s\n",
		lexer.CloseTag,
		node.Name)
}

func printTextNode(node TextNode) {
	utils.Debug("%s\t%s\n",
		lexer.Text,
		node.Data)
}

func printDynTextNode(node DynTextNode) {
	utils.Debug("%s\t%s\n",
		lexer.Code,
		node.Effect)
}

func printAttrNode(node AttrNode) {
	utils.Debug("%s\t%s\t%s\t%s\n",
		lexer.Attr,
		node.Name,
		node.Value,
		node.ValueType)
}

func printEventAttrNode(node EventAttrNode) {
	utils.Debug("%s\t%s\t%s\t%s\t%s\n",
		lexer.EventAttr,
		node.Name,
		node.Event,
		node.Effect,
		node.EffectType)
}

func (_self *Ast) processR(node Node[ElemNode]) {
	_self.ElemNodeProcessor(node.Value)
	for i := 0; i < len(node.Value.Children); i++ {
		var child = node.Value.Children[i]

		elemNode, ok := child.(Node[ElemNode])
		if ok {
			_self.processR(elemNode)
		}

		compNode, ok := child.(Node[CompNode])
		if ok {
			_self.CompNodeNodeProcessor(compNode.Value)
		}

		textNode, ok := child.(Node[TextNode])
		if ok {
			_self.TextNodeProcessor(textNode.Value)
		}

		dynTextNode, ok := child.(Node[DynTextNode])
		if ok {
			_self.DynTextNodeProcessor(dynTextNode.Value)
		}

		attrNode, ok := child.(Node[AttrNode])
		if ok {
			_self.AttrNodeProcessor(attrNode.Value)
		}

		eventAttrNode, ok := child.(Node[EventAttrNode])
		if ok {
			_self.EventAttrNodeProcessor(eventAttrNode.Value)
		}

		endNode, ok := child.(Node[EndNode])
		if ok {
			_self.EndNodeProcessor(endNode.Value)
		}
	}
}
