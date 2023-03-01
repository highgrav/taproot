package gfparser

import (
	"errors"
	"fmt"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/languages/token"
	"strings"
)

type ParserNodeType string

const (
	NODE_DOCUMENT  ParserNodeType = "doc"
	NODE_NOP       ParserNodeType = "nop"
	NODE_OUTPUT    ParserNodeType = "output"
	NODE_TAG       ParserNodeType = "tag"
	NODE_SPECIAL   ParserNodeType = "special"
	NODE_ATTRIBUTE ParserNodeType = "attribute"
	NODE_EOF       ParserNodeType = "eof"
	NODE_ERROR     ParserNodeType = "error"
	NODE_STRING    ParserNodeType = "string"
	NODE_ID        ParserNodeType = "id"
	NODE_NUMBER    ParserNodeType = "number"
)

type Parser struct {
	script string
	tokens []token.Token
}

func NewParser(tokens []token.Token) *Parser {
	p := &Parser{
		tokens: tokens,
	}
	return p
}

type ParseNode struct {
	NodeType ParserNodeType
	NodeName string
	Data     string
	Children []ParseNode
	Depth    int
}

func (pn *ParseNode) IsOfType(types []ParserNodeType) bool {
	for _, v := range types {
		if pn.NodeType == v {
			return true
		}
	}
	return false
}

type ParserState struct {
	OK     bool
	output strings.Builder
	Nodes  []ParseNode
	index  int
	tokens *[]token.Token
	script *string
}

func (parse *ParserState) current() token.Token {
	if parse.index >= len(*parse.tokens) {
		return token.Token{
			Type: token.TOKEN_EOF,
		}
	}
	tok := (*parse.tokens)[parse.index]
	return tok
}

func (parse *ParserState) next() token.Token {
	if parse.index >= len(*parse.tokens) {
		return token.Token{
			Type: token.TOKEN_EOF,
		}
	}
	tok := (*parse.tokens)[parse.index]
	parse.index++
	return tok
}

func (parse *ParserState) peek() token.Token {
	if parse.index+1 >= len(*parse.tokens) {
		return token.Token{
			Type: token.TOKEN_EOF,
		}
	}
	tok := (*parse.tokens)[parse.index+1]
	return tok
}

func (parse *ParserState) parse() error {
	depth := 0
	parse.OK = false
	parse.index = 0
	parse.Nodes = make([]ParseNode, 1)
	topNode := ParseNode{
		NodeType: NODE_DOCUMENT,
		Children: make([]ParseNode, 0),
	}

	currentNode, err := parse.parseElement(depth)

	for err == nil && currentNode.NodeType != NODE_ERROR && currentNode.NodeType != NODE_EOF {
		topNode.Children = append(topNode.Children, currentNode)
		currentNode, err = parse.parseElement(depth)
	}
	// handle error
	if err != nil {
		return err
	}

	if currentNode.NodeType == NODE_ERROR {
		return parse.throwError("")
	}

	parse.Nodes = append(parse.Nodes, topNode)
	parse.OK = true
	return nil
}

func (parse *ParserState) throwError(msg string) error {
	message := msg
	tok := parse.current()
	rs := common.NewRuneString(*parse.script)
	line, linepos := rs.GetLineAndPos(int32(tok.CharPos))
	if msg == "" {
		message = tok.Message
	}
	return errors.New(fmt.Sprintf("error '%s' at line %d, pos %d, context '%s'", message, line, linepos, tok.Literal))
}

func (parse *ParserState) parseElement(depth int) (ParseNode, error) {
	switch parse.current().Type {

	// both CDATA and TEXT are treated as textual output, and cannot have any children
	// STRINGS are also treated as output in this function, but not in the parseTag() function
	case token.TOKEN_SPECIAL_CDATA:
	case token.TOKEN_TEXT:
	case token.TOKEN_STRING:
		pt := ParseNode{
			NodeType: NODE_OUTPUT,
			NodeName: "cdata",
			Data:     parse.current().Literal,
			Children: make([]ParseNode, 0),
			Depth:    depth,
		}
		parse.next()
		return pt, nil
	// both close tags and comments generate NOPs
	case token.TOKEN_CLOSE_TAG:
	case token.TOKEN_END_SELF_CLOSING_TAG:
		pt := ParseNode{
			NodeType: NODE_NOP,
			NodeName: "closetag",
			Data:     parse.current().Literal[2 : len(parse.current().Literal)-1],
			Children: make([]ParseNode, 0),
			Depth:    depth,
		}
		parse.next()
		return pt, nil
	case token.TOKEN_SPECIAL_COMMENT:
		pt := ParseNode{
			NodeType: NODE_NOP,
			NodeName: "closetag",
			Data:     "",
			Children: make([]ParseNode, 0),
			Depth:    depth,
		}
		parse.next()
		return pt, nil
	case token.TOKEN_SPECIAL_OTHER:
		pt := ParseNode{
			NodeType: NODE_SPECIAL,
			NodeName: "special",
			Data:     parse.current().Literal,
			Children: make([]ParseNode, 0),
			Depth:    depth,
		}
		parse.next()
		return pt, nil
	case token.TOKEN_EOF:
		pt := ParseNode{
			NodeType: NODE_EOF,
			NodeName: "eof",
			Data:     "",
			Children: make([]ParseNode, 0),
			Depth:    depth,
		}
		parse.next()
		return pt, nil
	case token.TOKEN_ERROR:
		pt := ParseNode{
			NodeType: NODE_ERROR,
			NodeName: "error",
			Data:     parse.current().Literal,
			Children: make([]ParseNode, 0),
			Depth:    depth,
		}
		parse.next()
		return pt, nil
	case token.TOKEN_START_OPEN_TAG:
		pt, err := parse.parseTag(depth)
		return pt, err
	default:
		return ParseNode{}, nil
	}
	return ParseNode{}, nil
}

func (parse *ParserState) parseTag(depth int) (ParseNode, error) {
	currDepth := depth + 1
	if parse.current().Type != token.TOKEN_START_OPEN_TAG {
		return ParseNode{NodeType: "error"}, parse.throwError("tried to parse open tag on type " + string(parse.current().Type))
	}
	// Note that when we get the NodeName, we need to remove the prefixed '<'
	pt := ParseNode{
		NodeType: NODE_TAG,
		NodeName: string(parse.current().Literal[1:]),
		Data:     string(parse.current().Literal[1:]),
		Children: make([]ParseNode, 0),
	}

	parse.next()
	isInAssignment := false
	attrNode := ParseNode{
		Children: make([]ParseNode, 0),
	}
	for parse.current().IsOfType([]token.TokenType{token.TOKEN_ID, token.TOKEN_STRING, token.TOKEN_ASSIGN, token.TOKEN_NUMBER}) {
		// If we get a string or number without an assignment, we have a parse error
		if !isInAssignment && parse.current().Type == token.TOKEN_ID {
			if attrNode.NodeName != "" {
				// we have a bare ID (e.g., <checkbox checked />, so add the children and create a new one
				pt.Children = append(pt.Children, attrNode)
				attrNode = ParseNode{
					Children: make([]ParseNode, 0),
				}
			}
			attrNode.NodeType = NODE_ATTRIBUTE
			attrNode.NodeName = parse.current().Literal
			attrNode.Data = parse.current().Literal
		} else if !isInAssignment && parse.current().IsOfType([]token.TokenType{token.TOKEN_STRING, token.TOKEN_NUMBER}) {
			return ParseNode{
				NodeType: NODE_ERROR,
				NodeName: string(parse.current().Type),
				Data:     parse.current().Literal,
				Children: make([]ParseNode, 0),
			}, parse.throwError("found number of string outside of an attribute pair")
		} else if isInAssignment && parse.current().IsOfType([]token.TokenType{token.TOKEN_STRING, token.TOKEN_NUMBER, token.TOKEN_ID}) {
			nt := NODE_NOP
			if parse.current().Type == token.TOKEN_STRING {
				nt = NODE_STRING
			} else if parse.current().Type == token.TOKEN_NUMBER {
				nt = NODE_NUMBER
			} else if parse.current().Type == token.TOKEN_ID {
				nt = NODE_ID
			} else {
				return ParseNode{
					NodeType: NODE_ERROR,
					NodeName: string(parse.current().Type),
					Data:     parse.current().Literal,
					Children: make([]ParseNode, 0),
				}, parse.throwError("found something other than an id, number, or string in attribute assignment")
			}
			attrNode.Children = []ParseNode{ParseNode{
				NodeType: nt,
				NodeName: "value",
				Data:     parse.current().Literal,
				Children: nil,
				Depth:    0,
			}}
			pt.Children = append(pt.Children, attrNode)
			attrNode = ParseNode{
				Children: make([]ParseNode, 0),
			}
			isInAssignment = false
		} else if parse.current().Type == token.TOKEN_ASSIGN {
			isInAssignment = true
		}

		parse.next()
	}
	// append bareword IDs that otherwise wouldn't have been appended
	if attrNode.NodeType == NODE_ATTRIBUTE && len(attrNode.Children) == 0 {
		pt.Children = append(pt.Children, attrNode)
		attrNode = ParseNode{
			Children: make([]ParseNode, 0),
		}
	}

	// self-closing tags can end here
	if parse.current().Type == token.TOKEN_END_SELF_CLOSING_TAG {
		currDepth--
		parse.next()
		return pt, nil
	}

	// If this is a non-self-closing tag, then we close the tag but add children
	parse.next()
	// TODO
	return pt, nil // THIS IS WRONG

}
