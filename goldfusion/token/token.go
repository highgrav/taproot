package token

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
	TOKEN_TRUE                 TokenType = "true"
	TOKEN_FALSE                TokenType = "false"
	TOKEN_ASSIGN               TokenType = "assign"
	TOKEN_TEXT                 TokenType = "text"
	TOKEN_EOF                  TokenType = "eof"
	TOKEN_ERROR                TokenType = "error"
	TOKEN_SPECIAL_OTHER        TokenType = "specialother"
	TOKEN_SPECIAL_COMMENT      TokenType = "specialcomment"
	TOKEN_SPECIAL_CDATA        TokenType = "specialcdata"
)

type Token struct {
	tokenStartAt int
	Type         TokenType
	Literal      string
	Message      string
}

func New(tokenType TokenType, ch rune) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}
