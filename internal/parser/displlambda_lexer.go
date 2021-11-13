// Code generated from DisplLambda.g4 by ANTLR 4.9. DO NOT EDIT.

package parser

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 9, 51, 8,
	1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9,
	7, 4, 8, 9, 8, 3, 2, 3, 2, 3, 3, 3, 3, 3, 4, 3, 4, 3, 5, 3, 5, 3, 6, 6,
	6, 27, 10, 6, 13, 6, 14, 6, 28, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3,
	7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 8, 6, 8, 46, 10, 8, 13,
	8, 14, 8, 47, 3, 8, 3, 8, 2, 2, 9, 3, 3, 5, 4, 7, 5, 9, 6, 11, 7, 13, 8,
	15, 9, 3, 2, 4, 3, 2, 50, 59, 5, 2, 11, 12, 15, 15, 34, 34, 2, 52, 2, 3,
	3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2, 7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2, 2, 11,
	3, 2, 2, 2, 2, 13, 3, 2, 2, 2, 2, 15, 3, 2, 2, 2, 3, 17, 3, 2, 2, 2, 5,
	19, 3, 2, 2, 2, 7, 21, 3, 2, 2, 2, 9, 23, 3, 2, 2, 2, 11, 26, 3, 2, 2,
	2, 13, 30, 3, 2, 2, 2, 15, 45, 3, 2, 2, 2, 17, 18, 7, 44, 2, 2, 18, 4,
	3, 2, 2, 2, 19, 20, 7, 49, 2, 2, 20, 6, 3, 2, 2, 2, 21, 22, 7, 45, 2, 2,
	22, 8, 3, 2, 2, 2, 23, 24, 7, 47, 2, 2, 24, 10, 3, 2, 2, 2, 25, 27, 9,
	2, 2, 2, 26, 25, 3, 2, 2, 2, 27, 28, 3, 2, 2, 2, 28, 26, 3, 2, 2, 2, 28,
	29, 3, 2, 2, 2, 29, 12, 3, 2, 2, 2, 30, 31, 7, 116, 2, 2, 31, 32, 7, 103,
	2, 2, 32, 33, 7, 99, 2, 2, 33, 34, 7, 102, 2, 2, 34, 35, 7, 69, 2, 2, 35,
	36, 7, 85, 2, 2, 36, 37, 7, 88, 2, 2, 37, 38, 7, 78, 2, 2, 38, 39, 7, 107,
	2, 2, 39, 40, 7, 112, 2, 2, 40, 41, 7, 103, 2, 2, 41, 42, 7, 42, 2, 2,
	42, 43, 7, 43, 2, 2, 43, 14, 3, 2, 2, 2, 44, 46, 9, 3, 2, 2, 45, 44, 3,
	2, 2, 2, 46, 47, 3, 2, 2, 2, 47, 45, 3, 2, 2, 2, 47, 48, 3, 2, 2, 2, 48,
	49, 3, 2, 2, 2, 49, 50, 8, 8, 2, 2, 50, 16, 3, 2, 2, 2, 5, 2, 28, 47, 3,
	8, 2, 2,
}

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "'*'", "'/'", "'+'", "'-'", "", "'readCSVLine()'",
}

var lexerSymbolicNames = []string{
	"", "MUL", "DIV", "ADD", "SUB", "NUMBER", "READCSVLINE", "WHITESPACE",
}

var lexerRuleNames = []string{
	"MUL", "DIV", "ADD", "SUB", "NUMBER", "READCSVLINE", "WHITESPACE",
}

type DisplLambdaLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

// NewDisplLambdaLexer produces a new lexer instance for the optional input antlr.CharStream.
//
// The *DisplLambdaLexer instance produced may be reused by calling the SetInputStream method.
// The initial lexer configuration is expensive to construct, and the object is not thread-safe;
// however, if used within a Golang sync.Pool, the construction cost amortizes well and the
// objects can be used in a thread-safe manner.
func NewDisplLambdaLexer(input antlr.CharStream) *DisplLambdaLexer {
	l := new(DisplLambdaLexer)
	lexerDeserializer := antlr.NewATNDeserializer(nil)
	lexerAtn := lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)
	lexerDecisionToDFA := make([]*antlr.DFA, len(lexerAtn.DecisionToState))
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "DisplLambda.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// DisplLambdaLexer tokens.
const (
	DisplLambdaLexerMUL         = 1
	DisplLambdaLexerDIV         = 2
	DisplLambdaLexerADD         = 3
	DisplLambdaLexerSUB         = 4
	DisplLambdaLexerNUMBER      = 5
	DisplLambdaLexerREADCSVLINE = 6
	DisplLambdaLexerWHITESPACE  = 7
)
