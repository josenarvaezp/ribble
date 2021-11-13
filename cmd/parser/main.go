package main

import (
	"fmt"
	"os"

	"github.com/josenarvaezp/displ/internal/parser"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type TreeShapeListener struct {
	*parser.BaseDisplLambdaListener
}

func NewTreeShapeListener() *TreeShapeListener {
	return new(TreeShapeListener)
}

func (this *TreeShapeListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
	fmt.Println(ctx.GetText())
}

func main() {
	input, _ := antlr.NewFileStream(os.Args[1])
	lexer := parser.NewDisplLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewDisplParser(stream)
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.BuildParseTrees = true
	tree :=
		antlr.ParseTreeWalkerDefault.Walk(NewTreeShapeListener(), tree)
}
