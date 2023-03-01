package token

import "fmt"

type TokenType string

const (
	TOKEN_ILLEGAL              TokenType = "illegal"
	TOKEN_START                TokenType = "start"
	TOKEN_START_OPEN_TAG       TokenType = "startopentag"
	TOKEN_END_OPEN_TAG         TokenType = "endopentag"
	TOKEN_END_SELF_CLOSING_TAG TokenType = "endselfclosingtag"
	TOKEN_CLOSE_TAG            TokenType = "closetag"
	TOKEN_ID                   TokenType = "id"
	TOKEN_STRING               TokenType = "string"
	TOKEN_NUMBER               TokenType = "number"
	TOKEN_ASSIGN               TokenType = "assign"
	TOKEN_TEXT                 TokenType = "text"
	TOKEN_EOF                  TokenType = "eof"
	TOKEN_ERROR                TokenType = "error"
	TOKEN_SPECIAL_OTHER        TokenType = "specialother"
	TOKEN_SPECIAL_COMMENT      TokenType = "specialcomment"
	TOKEN_SPECIAL_CDATA        TokenType = "specialcdata"
)

type Token struct {
	CharPos int
	LinePos int
	Type    TokenType
	Literal string
	Message string
}

func (pn Token) Dump() string {
	return fmt.Sprintf("Type: %s, Message: %s, Literal: %s", pn.Type, pn.Message, pn.Literal)
}

func (pn Token) IsNotOfType(types []TokenType) bool {
	for _, v := range types {
		if pn.Type == v {
			return false
		}
	}
	return true
}

func (pn Token) IsOfType(types []TokenType) bool {
	for _, v := range types {
		if pn.Type == v {
			return true
		}
	}
	return false
}

func New(tokenType TokenType, ch rune) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}
