package lexer

import (
	"errors"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/languages/token"
)

const (
	ATOM_LT     rune = '<'
	ATOM_RT     rune = '>'
	ATOM_ASSIGN rune = '='
	ATOM_DQOUT  rune = '"'
	ATOM_SQUOTE rune = '\''
	ATOM_BANG   rune = '!'
	ATOM_DASH   rune = '-'
	ATOM_SLASH  rune = '/'
)

func isDelimiter(ch rune, delimiters []rune) bool {
	for _, c := range delimiters {
		if ch == c {
			return true
		}
	}
	return false
}

func isLetter(ch rune) bool {
	return rune('a') <= ch && ch <= rune('z') || rune('A') <= ch && ch <= rune('Z') || ch == rune('_')
}

func isIDLetter(ch rune) bool {
	return rune('a') <= ch && ch <= rune('z') || rune('A') <= ch && ch <= rune('Z') || ch == rune('_') || ch == rune('.') || ch == rune('-') || ch == rune('@')
}

func isNumber(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}

func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('	') || ch == rune('\r') || ch == rune('\n')
}

type Lexer struct {
	input        *common.RuneString
	position     int32
	readPosition int32
	totalLines   int32
	ch           rune
}

func New(script string) *Lexer {
	l := &Lexer{
		input:        common.NewRuneString(script),
		position:     0,
		readPosition: 0,
		totalLines:   0,
		ch:           0,
	}
	// Warm up the lexer
	l.readChar()

	return l
}

func (lex *Lexer) Peek() (rune, bool) {
	if (lex.readPosition) >= lex.input.Length {
		return rune(0x0), false
	}
	return lex.input.Peek(), true
}

func (lex *Lexer) advanceToNextNonWhitespaceChar() {
	for isWhitespace(lex.ch) {
		if lex.ch == rune('\n') {
			lex.totalLines = lex.totalLines + 1
		}
		lex.readChar()
	}
}

func (lex *Lexer) readChar() {
	lex.ch = lex.input.Get(lex.readPosition)
	lex.position = lex.readPosition
	lex.readPosition += 1
}

func (lex *Lexer) readIdentifier() string {
	pos := lex.position
	for (isIDLetter(lex.ch) && lex.ch != rune(0x0)) || (isNumber(lex.ch) && lex.position != pos) {
		lex.readChar()
	}
	return string(lex.input.Runes[pos:lex.position])
}

func (lex *Lexer) readNumber() string {
	pos := lex.position
	seenDecimal := false
	for (isNumber(lex.ch) && lex.ch != rune(0)) || (lex.ch == rune('.') && !seenDecimal) {
		if lex.ch == rune('.') {
			seenDecimal = true
		}
		lex.readChar()
	}
	return string(lex.input.Runes[pos:lex.position])
}

func (lex *Lexer) readDelimitedString(delimiter rune) string {
	pos := lex.position
	isEscaped := false
	for (lex.ch != delimiter || pos == lex.position) && lex.ch != rune(0) {
		if !isEscaped && lex.ch == rune('\\') {
			isEscaped = true
		} else if isEscaped {
			isEscaped = false
		}
		lex.readChar()
	}

	// TODO -- escaped characters?
	return string(lex.input.Runes[pos:lex.readPosition])
}

func (lex *Lexer) readTextToDelimiter(delimiters []rune) string {
	var txt string = ""
	for lex.ch != rune(0x0) && !isDelimiter(lex.ch, []rune{'<', '"', '\''}) {
		if lex.ch == '\n' {
			lex.totalLines++
		}
		txt = txt + string(lex.ch)
		lex.readChar()
	}
	return txt
}

// Tags are defined by starting with <[A-Za-z_], </, or <!
func (lex *Lexer) readTag() []token.Token {
	if isLetter(lex.input.Get(lex.readPosition)) {
		return lex.readOpenTag()
	} else if lex.input.Get(lex.readPosition) == ATOM_SLASH {
		return lex.readCloseTag()
	} else if lex.input.Get(lex.readPosition) == ATOM_BANG {
		return lex.readOpenMarkup()
	} else {
		return []token.Token{token.Token{
			Type:    token.TOKEN_TEXT,
			Literal: string(lex.ch),
			CharPos: int(lex.position),
		}}
	}
}

func (lex *Lexer) readOpenTag() []token.Token {
	tokens := make([]token.Token, 0)
	if lex.ch != '<' {
		return []token.Token{
			token.Token{
				Type:    token.TOKEN_ERROR,
				Literal: string(lex.ch),
				Message: "attempted to read opening tag but got unexpected prefix",
				CharPos: int(lex.position),
			},
		}
	}
	if lex.input.Get(lex.readPosition) == '!' {
		return lex.readOpenMarkup()
	}
	lit := string(lex.ch)
	lex.readChar()
	id := lex.readIdentifier()
	lit = lit + id
	tokens = append(tokens, token.Token{
		Type:    token.TOKEN_START_OPEN_TAG,
		Literal: lit,
		Message: "",
		CharPos: int(lex.position),
	})

	// walk through the string until we see an ending character
	for lex.ch != '/' && lex.ch != '>' {
		if lex.ch == '"' || lex.ch == '\'' {
			ttoks := lex.readString()
			for _, t := range ttoks {
				tokens = append(tokens, t)
			}
			// note that we need to advance past the end quote
			lex.readChar()
		} else if isNumber(lex.ch) {
			tokens = append(tokens, token.Token{
				Type:    token.TOKEN_NUMBER,
				Literal: lex.readNumber(),
				Message: "",
				CharPos: int(lex.position),
			})
		} else if lex.ch == '=' {
			tokens = append(tokens, token.Token{
				Type:    token.TOKEN_ASSIGN,
				Literal: string(lex.ch),
				Message: "",
				CharPos: int(lex.position),
			})
			lex.readChar()
		} else if isWhitespace(lex.ch) {
			// NOP
			lex.advanceToNextNonWhitespaceChar()
		} else if lex.ch == rune(0x0) {
			tokens = append(tokens, token.Token{
				Type:    token.TOKEN_ERROR,
				Literal: string(lex.ch),
				Message: "unexpected eof detected while parsing open tag",
				CharPos: int(lex.position),
			})
			return tokens
		} else {
			tokens = append(tokens, token.Token{
				Type:    token.TOKEN_ID,
				Literal: lex.readIdentifier(),
				Message: "",
				CharPos: int(lex.position),
			})
		}
		lex.advanceToNextNonWhitespaceChar()
	}

	if lex.ch == '>' {
		lit = lit + string(lex.ch)
		tokens = append(tokens, token.Token{
			Type:    token.TOKEN_END_OPEN_TAG,
			Literal: lit,
			Message: "",
			CharPos: int(lex.position),
		})
		lex.readChar()
	} else if lex.ch == '/' && lex.input.Get(lex.readPosition) == '>' {
		lit = lit + string(lex.ch)
		lex.readChar()
		lit = lit + string(lex.ch)
		tokens = append(tokens, token.Token{
			Type:    token.TOKEN_END_SELF_CLOSING_TAG,
			Literal: lit,
			Message: "",
			CharPos: int(lex.position),
		})
		lex.readChar()
	} else {
		lit = lit + string(lex.ch)
		tokens = append(tokens, token.Token{
			Type:    token.TOKEN_ERROR,
			Literal: lit,
			Message: "attempted to close open tag but got unexpected suffix",
			CharPos: int(lex.position),
		})
	}
	return tokens
}

// TODO -- this is a bit fragile right now, since it naively reads to an '>'
func (lex *Lexer) readCloseTag() []token.Token {
	if lex.ch != '<' && lex.input.Get(lex.readPosition) != '/' {
		return []token.Token{
			token.Token{
				Type:    token.TOKEN_ERROR,
				Literal: string(lex.ch + lex.input.Get(lex.readPosition)),
				Message: "attempted to read closing tag but got unexpected prefix",
				CharPos: int(lex.position),
			},
		}
	}
	lit := ""
	for lex.ch != '>' {
		lit = lit + string(lex.ch)
		lex.readChar()
	}
	lit = lit + string(lex.ch)
	lex.readChar()
	return []token.Token{
		token.Token{
			Type:    token.TOKEN_CLOSE_TAG,
			Literal: lit,
			CharPos: int(lex.position),
		},
	}
}

// Open markup starts with <!, and can consist of:
// A comment <!-- ... -->
// CDATA <![CDATA[ ... ]]>
// And the <![A-Za-z] types:
//
//	Doctype <!DOCTYPE ... >
//	Element <!ELEMENT ... >
//	Attlist <!ATTLIST ... >
//
// Open markup is ignored right now, and shoved into a generic "open markup" token
func (lex *Lexer) readOpenMarkup() []token.Token {
	tokens := make([]token.Token, 0)
	if lex.ch != '<' {
		lex.readChar()
		return []token.Token{
			token.Token{
				Type:    token.TOKEN_ERROR,
				Literal: string(lex.ch),
				Message: "attempted to read open markup tag but got unexpected prefix (expected <)",
				CharPos: int(lex.position),
			},
		}
	}
	if lex.input.Get(lex.readPosition) != '!' {
		lex.readChar()
		lex.readChar()
		return []token.Token{
			token.Token{
				Type:    token.TOKEN_ERROR,
				Literal: string(lex.ch),
				Message: "attempted to read open markup tag but got unexpected prefix (expected <!)",
				CharPos: int(lex.position),
			},
		}
	}
	str := string(lex.ch)
	lex.readChar()
	str = str + string(lex.ch)
	lex.readChar()
	if lex.input.MatchesFrom(lex.position, "--") {
		// Comment!
		for !lex.input.MatchesFrom(lex.position, "-->") {
			str = str + string(lex.ch)
			lex.readChar()
		}

		// read out -->
		lex.readChar() // -
		str = str + string(lex.ch)

		lex.readChar() // -
		str = str + string(lex.ch)

		lex.readChar() // >
		str = str + string(lex.ch)

		lex.readChar()
		tokens = append(tokens, token.Token{
			Type:    token.TOKEN_SPECIAL_COMMENT,
			Literal: str,
			Message: "",
			CharPos: int(lex.position),
		})
		return tokens
	} else if lex.input.MatchesFrom(lex.position, "[CDATA[") {
		// With CDATA, we strip out the CDATA prefix and suffix, so it looks like a pure output
		lex.readChar()
		lex.readChar()
		lex.readChar()
		lex.readChar()
		lex.readChar()
		lex.readChar()
		lex.readChar()
		str = str[:len(str)-2] // remove the <!
		for !lex.input.MatchesFrom(lex.position, "]]>") {
			str = str + string(lex.ch)
			lex.readChar()
		}

		lex.readChar() // ]
		lex.readChar() // ]
		lex.readChar() // >
		lex.readChar()
		tokens = append(tokens, token.Token{
			Type:    token.TOKEN_SPECIAL_CDATA,
			Literal: str,
			Message: "",
			CharPos: int(lex.position),
		})
		return tokens
	} else {
		// TODO -- this is very naive, may need to replace it with tag-like functionality

		for lex.ch != '>' {
			if lex.ch == rune(0x0) {
				tokens = append(tokens, token.Token{
					Type:    token.TOKEN_ERROR,
					Literal: str,
					Message: "unexpected eof while reading open markup tag",
					CharPos: int(lex.position),
				})
				return tokens
			}
			str = str + string(lex.ch)
			lex.readChar()
		}
		str = str + string(lex.ch)
		lex.readChar()
		tokens = append(tokens, token.Token{
			Type:    token.TOKEN_SPECIAL_OTHER,
			Literal: str,
			Message: "",
			CharPos: int(lex.position),
		})
		return tokens
	}
	return tokens
}

func (lex *Lexer) readText() []token.Token {
	// Put the first character in the string, since it may be a <
	str := string(lex.ch)
	lex.readChar()
	str = str + lex.readTextToDelimiter([]rune{'"', '\'', '<'})
	return []token.Token{token.Token{
		Type:    token.TOKEN_TEXT,
		Literal: str,
		CharPos: int(lex.position),
	}}
}

func (lex *Lexer) readString() []token.Token {
	str := lex.readDelimitedString(lex.ch)
	return []token.Token{token.Token{
		Type:    token.TOKEN_STRING,
		Literal: str,
		CharPos: int(lex.position),
	}}
}

func (lex *Lexer) NextToken() []token.Token {
	var toks []token.Token
	if lex.ch == rune(0x0) {
		t := token.Token{}
		t.Type = token.TOKEN_EOF
		t.CharPos = int(lex.position)
		return []token.Token{t}
	} else if lex.ch == ATOM_LT {

		// check to see if the next character is a letter; if so, read tag, otherwise assume text
		if isLetter(lex.input.Get(lex.readPosition)) || lex.input.Get(lex.readPosition) == ATOM_SLASH || lex.input.Get(lex.readPosition) == ATOM_BANG {
			// read as tag
			toks = lex.readTag()
		} else {
			// read as text
			toks = lex.readText()
		}
	} else if lex.ch == ATOM_SQUOTE || lex.ch == ATOM_DQOUT {
		// read string
		toks = lex.readString()
	} else {
		// read text
		toks = lex.readText()
	}
	return toks
}

func (lex *Lexer) Lex() ([]token.Token, error) {
	var tokens []token.Token = make([]token.Token, 0)

	toks := lex.NextToken()
	for true {
		for _, tok := range toks {
			if tok.Type == token.TOKEN_EOF {
				tokens = append(tokens, tok)
				return tokens, nil
			}
			if tok.Type == token.TOKEN_ERROR {
				tokens = append(tokens, tok)
				return tokens, errors.New(tok.Message)
			}
			tokens = append(tokens, tok)
		}
		toks = lex.NextToken()
	}
	return tokens, nil
}
