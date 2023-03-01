package naive

import (
	html "golang.org/x/net/html"
	"io"
)

type tokenStream struct {
	tokenNodes []html.Token
}

const (
	VariableOutputToken html.TokenType = 3232
)

// This does a simple high-level pass over a gfm page and attempts to break it down to tokens.
// This works for most but not all cases; some tags do not allow nested tags, so when we find
// a text node we need to use tokenizeString() to break it into any component tokens.
func tokenizeHtml(pageReader io.Reader) ([]html.Token, error) {
	tokenizer := html.NewTokenizer(pageReader)
	tokenNodes := make([]html.Token, 0)
	var tokenType html.TokenType = tokenizer.Next()
	for ; tokenType != html.ErrorToken; tokenType = tokenizer.Next() {

		switch tokenType {
		case html.DoctypeToken:
			tokenNodes = append(tokenNodes, tokenizer.Token())
			continue
		case html.CommentToken:
			tokenNodes = append(tokenNodes, tokenizer.Token())
			continue
		case html.StartTagToken:
			tokenNodes = append(tokenNodes, tokenizer.Token())
			continue
		case html.EndTagToken:
			tokenNodes = append(tokenNodes, tokenizer.Token())
			continue
		case html.SelfClosingTagToken:
			tokenNodes = append(tokenNodes, tokenizer.Token())
			continue
		case html.TextToken:
			tns, err := tokenizeString(tokenizer.Token())
			if err != nil {
				return tokenNodes, err
			}
			for _, tn := range tns {
				tokenNodes = append(tokenNodes, tn)
			}
			continue
		default:
			continue

		}
	}
	if tokenizer.Err() != nil && tokenizer.Err() != io.EOF {
		return tokenNodes, tokenizer.Err()
	}

	return tokenNodes, nil
}

func tokenizeString(token html.Token) ([]html.Token, error) {
	// TODO
	return []html.Token{token}, nil
}
