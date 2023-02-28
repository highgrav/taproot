package parser

import "highgrav/taproot/v1/goldfusion/token"

type Parser struct {
	tokens []token.Token
}

type ParseNode struct {
}

func NewParser(tokens []token.Token) *Parser {
	p := &Parser{
		tokens: tokens,
	}
	return p
}
