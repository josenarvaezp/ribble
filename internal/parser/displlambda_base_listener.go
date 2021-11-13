// Code generated from DisplLambda.g4 by ANTLR 4.9. DO NOT EDIT.

package parser // DisplLambda

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseDisplLambdaListener is a complete listener for a parse tree produced by DisplLambdaParser.
type BaseDisplLambdaListener struct{}

var _ DisplLambdaListener = &BaseDisplLambdaListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseDisplLambdaListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseDisplLambdaListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseDisplLambdaListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseDisplLambdaListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterStart is called when production start is entered.
func (s *BaseDisplLambdaListener) EnterStart(ctx *StartContext) {}

// ExitStart is called when production start is exited.
func (s *BaseDisplLambdaListener) ExitStart(ctx *StartContext) {}

// EnterNumber is called when production Number is entered.
func (s *BaseDisplLambdaListener) EnterNumber(ctx *NumberContext) {}

// ExitNumber is called when production Number is exited.
func (s *BaseDisplLambdaListener) ExitNumber(ctx *NumberContext) {}

// EnterMulDiv is called when production MulDiv is entered.
func (s *BaseDisplLambdaListener) EnterMulDiv(ctx *MulDivContext) {}

// ExitMulDiv is called when production MulDiv is exited.
func (s *BaseDisplLambdaListener) ExitMulDiv(ctx *MulDivContext) {}

// EnterAddSub is called when production AddSub is entered.
func (s *BaseDisplLambdaListener) EnterAddSub(ctx *AddSubContext) {}

// ExitAddSub is called when production AddSub is exited.
func (s *BaseDisplLambdaListener) ExitAddSub(ctx *AddSubContext) {}
