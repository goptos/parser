package view

import (
	"fmt"
	"strings"

	"github.com/goptos/parser/ast"
	"github.com/goptos/parser/lexer"
	"github.com/goptos/utils"
)

const (
	xHtmlElem     int = 0
	HtmlEventAttr int = 1
	xHtmlAttr     int = 2
	HtmlText      int = 3
	HtmlDynText   int = 100
)

type Ast = ast.Ast
type TokenType = lexer.TokenType
type Token = lexer.Token

func View(source string) (*string, error) {
	var document, err = (*Ast).New(nil, source)
	if err != nil {
		utils.Debug("%s\n", err)
		return nil, err
	}
	var statements []string = make([]string, 0)
	var index int = 0
	/*
		`<div>` => `(*Elem).New(nil, "div")`
	*/
	document.ElemNodeProcessor = func(node ast.ElemNode) {
		utils.Debug("%d\t%s\t%s\n",
			index,
			lexer.OpenTag,
			node.Name)
		statements = append(statements, fmt.Sprintf(
			"(*Elem).New(nil, \"%s\")",
			node.Name))
		index = len(statements) - 1

		// check if we are a generating node
		/*
			`<ul each={cF} key={kF}><Li /></ul>` => `system.Each((*Elem).New(nil, "ul"), cx, cF, kF, Li.View)`
		*/
		var hasCollectF = false
		var cF = ""
		var hasKeyF = false
		var kF = ""
		var isEach = false
		var vF = ""
		for _, child := range node.Children {
			attrNode, ok := child.(ast.Node[ast.AttrNode])
			if ok {
				if !hasCollectF && attrNode.Value.Name == "each" {
					hasCollectF = true
					cF = attrNode.Value.Value
				}
				if hasCollectF && attrNode.Value.Name == "key" {
					hasKeyF = true
					kF = attrNode.Value.Value
				}
			}
			compNode, ok := child.(ast.Node[ast.CompNode])
			if hasKeyF && ok {
				isEach = true
				vF = compNode.Value.Name
			}
		}
		if isEach {
			statements[index] = "system.Each(" +
				statements[index] + ", cx, " +
				cF[1:len(cF)-1] + ", " +
				kF[1:len(kF)-1] + ", " +
				vF + ".View)"
		}
	}
	/*
		`<Button />` => `.Child(Button.View(cx))`
	*/
	document.CompNodeNodeProcessor = func(node ast.CompNode) {
		utils.Debug("%d\t%s\t%s\t%s\n",
			index,
			lexer.Comp,
			node.Name,
			strings.Join(node.Args, "\t"))
		// if this compNode was used as a vF in the preceding Each then do not process again
		if index-1 <= 0 {
			if strings.Contains(statements[index], ", "+node.Name+".View)") {
				return
			}
		}
		statements[index] = statements[index] + ".Child(" + node.Name + ".View(cx"
		if len(node.Args) > 0 {
			statements[index] = statements[index] + ", " + strings.Join(node.Args, ", ")
		}
		statements[index] = statements[index] + "))"
	}
	/*
		`Hello` => `Text("Hello")`
	*/
	document.TextNodeProcessor = func(node ast.TextNode) {
		utils.Debug("%d\t%s\t%s\n",
			index,
			lexer.Text,
			node.Data)
		statements[index] = statements[index] + fmt.Sprintf(
			".Text(\"%s\")",
			node.Data)
	}
	/*
		DynText(cx, func() string {
			return fmt.Sprintf("%s", count.Get())
		})
	*/
	document.DynTextNodeProcessor = func(node ast.DynTextNode) {
		utils.Debug("%d\t%s\t%s\n",
			index,
			lexer.Text,
			node.Effect)
		statements[index] = statements[index] + fmt.Sprintf(
			".DynText(cx, func() string { return fmt.Sprintf(\"%%v\", %s) })",
			node.Effect[1:len(node.Effect)-1])
	}
	/*
		Attr("id", "sub-button")
	*/
	document.AttrNodeProcessor = func(node ast.AttrNode) {
		utils.Debug("%d\t%s\t%s\t%s\t%s\n",
			index,
			lexer.Attr,
			node.Name,
			node.Value,
			node.ValueType)
		switch node.ValueType {
		case lexer.Code:
			// ToDo
		case lexer.Quote:
			statements[index] = statements[index] + fmt.Sprintf(
				".Attr(\"%s\", %s)",
				node.Name,
				node.Value)
		}
	}
	/*
		On("click", func(Event) {} )
	*/
	document.EventAttrNodeProcessor = func(node ast.EventAttrNode) {
		utils.Debug("%d\t%s\t%s\t%s\t%s\t%s\n",
			index,
			lexer.Attr,
			node.Name,
			node.Event,
			node.Effect,
			node.EffectType)
		statements[index] = statements[index] + fmt.Sprintf(
			".On(\"%s\", %s)",
			node.Event,
			node.Effect[1:len(node.Effect)-1])
	}
	/*
		</...>
	*/
	document.EndNodeProcessor = func(node ast.EndNode) {
		utils.Debug("%d\t%s\t%s\n",
			index,
			lexer.CloseTag,
			node.Name)
		if index-1 >= 0 {
			// check if we need to be a DynChild
			var condition = ""
			var isDynChild = false
			for _, child := range node.Elem.Children {
				attrNode, ok := child.(ast.Node[ast.AttrNode])
				if ok {
					if attrNode.Value.Name == "if" {
						isDynChild = true
						condition = attrNode.Value.Value
						break
					}
				}
			}
			if isDynChild {
				statements[index-1] = statements[index-1] + fmt.Sprintf(
					".DynChild(cx, %s, %s)",
					condition[1:len(condition)-1],
					statements[index])
			} else {
				statements[index-1] = statements[index-1] + fmt.Sprintf(
					".Child(%s)",
					statements[index])
			}
			statements = statements[0:index]
			index = len(statements) - 1
		}
	}
	document.Process()
	utils.Debug("\n\n%s\n\n", statements[0])
	return &statements[0], nil
}
