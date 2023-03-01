package jsmlparser

import (
	"errors"
	"fmt"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/languages/token"
	"strings"
)

var ErrorEOF error = errors.New("reached EOF")

type ParserNodeType string

const (
	NODE_DOCUMENT           ParserNodeType = "doc"
	NODE_NOP                ParserNodeType = "nop"
	NODE_OUTPUT             ParserNodeType = "output"
	NODE_TAG                ParserNodeType = "tag"
	NODE_SELF_CLOSE_TAG     ParserNodeType = "selfclosedtag"
	NODE_CLOSE_TAG          ParserNodeType = "closetag"
	NODE_SPECIAL_OTHER      ParserNodeType = "special"
	NODE_ATTRIBUTE          ParserNodeType = "attribute"
	NODE_EOF                ParserNodeType = "eof"
	NODE_ERROR              ParserNodeType = "error"
	NODE_STRING             ParserNodeType = "string"
	NODE_ID                 ParserNodeType = "id"
	NODE_NUMBER             ParserNodeType = "number"
	NODE_INTERPOLATED_VALUE ParserNodeType = "interpval"
)

type ParseNode struct {
	NodeType         ParserNodeType
	NodeName         string
	Data             string
	IsSelfClosingTag bool
	Children         []ParseNode
	Parent           *ParseNode
	Token            token.Token
}

func newNode(nType ParserNodeType, name string, data string, parent *ParseNode, tok token.Token) ParseNode {
	return ParseNode{
		NodeType:         nType,
		NodeName:         name,
		Data:             data,
		IsSelfClosingTag: false,
		Children:         make([]ParseNode, 0),
		Parent:           parent,
		Token:            tok,
	}
}

func (pn *ParseNode) IsOfType(types []ParserNodeType) bool {
	for _, v := range types {
		if pn.NodeType == v {
			return true
		}
	}
	return false
}

type Parser struct {
	Parsed bool
	script string
	tokens *[]token.Token
	tree   *ParseNode
	index  int
}

func New(toks *[]token.Token, script string) *Parser {
	return &Parser{
		Parsed: false,
		script: script,
		tokens: toks,
		tree:   nil,
		index:  0, // note that the index always points one ahead
	}
}

func (parse *Parser) Tree() *ParseNode {
	return parse.tree
}

func (parse *Parser) throwError(msg string) error {
	message := msg
	tok := parse.current()
	rs := common.NewRuneString(parse.script)
	line, linepos := rs.GetLineAndPos(int32(tok.CharPos))
	if msg == "" {
		message = tok.Message
	}
	return errors.New(fmt.Sprintf("error '%s' at line %d, pos %d, (token type: %s, literal: '%s')", message, line, linepos, tok.Type, tok.Literal))
}

func (parse *Parser) current() token.Token {
	if parse.index >= len(*parse.tokens) {
		return token.Token{
			Type: token.TOKEN_EOF,
		}
	}
	tok := (*parse.tokens)[parse.index]
	return tok
}

func (parse *Parser) next() token.Token {
	if parse.index >= len(*parse.tokens) {
		return token.Token{
			Type: token.TOKEN_EOF,
		}
	}
	tok := (*parse.tokens)[parse.index]
	parse.index++
	return tok
}

func (parse *Parser) Parse() error {
	parse.index = 0
	parse.tree = &ParseNode{
		NodeType: NODE_DOCUMENT,
		NodeName: "",
		Data:     "",
		Children: make([]ParseNode, 0),
		Parent:   nil,
		Token:    parse.current(),
	}
	depth := 0
	err := parse.parseElement(parse.tree)
	for err == nil {
		fmt.Println(depth)
		depth++
		err = parse.parseElement(parse.tree)
	}
	if errors.Is(err, ErrorEOF) {
		parse.Parsed = true
		return nil
	} else {
		return err
	}
}

// NOTE: The parse* functions are not idempotent and will mutate the node passed to them.
func (parse *Parser) parseElement(node *ParseNode) error {
	parse.next() // advance one
	switch parse.current().Type {
	case token.TOKEN_SPECIAL_CDATA:
		return parse.parseSpecialCData(node)
	case token.TOKEN_TEXT:
		return parse.parseText(node)
	case token.TOKEN_STRING:
		return parse.parseString(node)
	case token.TOKEN_SPECIAL_COMMENT:
		return parse.parseSpecialComment(node)
	case token.TOKEN_SPECIAL_OTHER:
		return parse.parseSpecialOtherTag(node)
	case token.TOKEN_EOF:
		node.Children = append(node.Children, newNode(NODE_EOF, "", "", node, parse.current()))
		return ErrorEOF
	case token.TOKEN_START_OPEN_TAG:
		if parse.current().Literal == "go" || strings.HasPrefix(parse.current().Literal, "go.") {
			return parse.parseSemanticTag(node)
		} else {
			return parse.parseNonSemanticTag(node)
		}
	case token.TOKEN_CLOSE_TAG:
		return parse.parseCloseTag(node)
	default:
		// how did we get here?
		return parse.throwError("unexpected token type")
	}
}

func (parse *Parser) parseNonSemanticTag(node *ParseNode) error {
	return parse.parseTag(node, false)
}

func (parse *Parser) parseSemanticTag(node *ParseNode) error {
	return parse.parseTag(node, true)
}

// Takes a start tag and parses it, its attributes, its children, and recursively its tag children, until finding a close tag
func (parse *Parser) parseTag(node *ParseNode, isSemantic bool) error {
	// parse tag start -- if we don't see a start to an open tag, we've got an error.
	if parse.current().Type != token.TOKEN_START_OPEN_TAG {
		pn := newNode(NODE_ERROR, "error", parse.current().Literal, node, parse.current())
		node.Children = append(node.Children, pn)
		return parse.throwError("tried to parse open tag on type " + string(parse.current().Type))
	}
	tagNode := newNode(NODE_TAG, string(parse.current().Literal[1:]), string(parse.current().Literal[1:]), node, parse.current())

	// parse attributes -- we should see either an ID or and ID ASSIGN [ID,STRING,NUMBER]
	isInAssignment := false
	attrNode := ParseNode{
		NodeType: NODE_ATTRIBUTE,
		Children: make([]ParseNode, 0),
		Token:    parse.current(),
		Parent:   node,
	}
	parse.next()
	for parse.current().IsOfType([]token.TokenType{token.TOKEN_ID, token.TOKEN_STRING, token.TOKEN_ASSIGN, token.TOKEN_NUMBER}) {
		if !isInAssignment && parse.current().Type == token.TOKEN_ID {
			if attrNode.NodeName != "" {
				tagNode.Children = append(tagNode.Children, attrNode)
				attrNode = ParseNode{
					NodeType: NODE_ATTRIBUTE,
					Children: make([]ParseNode, 0),
					Token:    parse.current(),
					Parent:   &tagNode,
				}
			}
			attrNode.NodeName = parse.current().Literal
			attrNode.Data = parse.current().Literal
		} else if !isInAssignment && parse.current().IsOfType([]token.TokenType{token.TOKEN_STRING, token.TOKEN_NUMBER}) {
			// this is an error
			node.Children = append(node.Children, tagNode)
			return parse.throwError("unexpected token while parsing tag header (should be ID)")
		} else if isInAssignment && parse.current().IsOfType([]token.TokenType{token.TOKEN_STRING, token.TOKEN_NUMBER, token.TOKEN_ID}) {
			valnode := newNode(NODE_NOP, string(parse.current().Type), parse.current().Literal, nil, parse.current())
			switch parse.current().Type {
			case token.TOKEN_ID:
				if strings.HasPrefix(parse.current().Literal, "@") {
					valnode.NodeType = NODE_INTERPOLATED_VALUE
				} else {
					valnode.NodeType = NODE_ID
				}
			case token.TOKEN_STRING:
				valnode.NodeType = NODE_STRING
			case token.TOKEN_NUMBER:
				valnode.NodeType = NODE_NUMBER
			default:
				node.Children = append(node.Children, tagNode)
				return parse.throwError("unexpected rval for attribute (should be ID, string, or number)")
			}
			valnode.Parent = &attrNode
			attrNode.Children = append(attrNode.Children, valnode)
			isInAssignment = false
		} else if parse.current().Type == token.TOKEN_ASSIGN {
			isInAssignment = true
		} else {
			node.Children = append(node.Children, tagNode)
			return parse.throwError("unexpected token while parsing tag header")
		}
		parse.next()
	}
	if attrNode.NodeName != "" {
		tagNode.Children = append(tagNode.Children, attrNode)
	}
	attrNode = newNode(NODE_NOP, "", "", nil, parse.current())

	// parse self-closing tag and return
	if parse.current().Type == token.TOKEN_END_SELF_CLOSING_TAG {
		tagNode.IsSelfClosingTag = true
		node.Children = append(node.Children, tagNode)
		return nil
	} else if parse.current().Type == token.TOKEN_END_OPEN_TAG {
		if !isSemantic {
			tagNode.IsSelfClosingTag = false
			node.Children = append(node.Children, tagNode)
			return nil
		}
	} else if parse.current().Type == token.TOKEN_EOF {
		node.Children = append(node.Children, tagNode)
		pn := newNode(NODE_EOF, "", "", node, parse.current())
		node.Children = append(node.Children, pn)
		return ErrorEOF
	} else if parse.current().Type == token.TOKEN_ERROR {
		pn := newNode(NODE_ERROR, "error", parse.current().Literal, node, parse.current())
		tagNode.Parent = node
		node.Children = append(node.Children, tagNode)
		node.Children = append(node.Children, pn)
		return parse.throwError("")
	} else {
		pn := newNode(NODE_ERROR, "error", parse.current().Literal, node, parse.current())
		tagNode.Parent = node
		node.Children = append(node.Children, tagNode)
		node.Children = append(node.Children, pn)
		return parse.throwError("failed to find close tag, instead got " + string(parse.current().Type))
	}

	// parse child elements until we see a close tag
	parse.next()
	for parse.current().IsNotOfType([]token.TokenType{token.TOKEN_EOF, token.TOKEN_ERROR}) {
		if parse.current().Type == token.TOKEN_CLOSE_TAG && parse.current().Literal == "</"+tagNode.Data+">" {
			break
		}
		err := parse.parseElement(&tagNode)
		if err != nil {
			node.Children = append(node.Children, tagNode)
			return err
		}
		parse.next()
	}

	if parse.current().Type == token.TOKEN_CLOSE_TAG {
		// SPECIAL CASE -- only return if the close tag matches the current top-level tag
		err := parse.parseCloseTag(&tagNode)
		if err != nil {
			pn := newNode(NODE_ERROR, "error", parse.current().Literal, node, parse.current())
			node.Children = append(node.Children, pn)
			return parse.throwError("")
		}
		node.Children = append(node.Children, tagNode)
		return nil

	} else if parse.current().Type == token.TOKEN_ERROR {
		node.Children = append(node.Children, tagNode)
		pn := newNode(NODE_ERROR, "error", parse.current().Literal, node, parse.current())
		node.Children = append(node.Children, pn)
		return parse.throwError("")
	} else if parse.current().Type == token.TOKEN_EOF {
		node.Children = append(node.Children, tagNode)
		node.Children = append(node.Children, newNode(NODE_EOF, "", "", node, parse.current()))
		return ErrorEOF
	}
	return parse.throwError("unknown syntax error, fallthrough while parsing tag recursion")
}

// A string outside of a tag (proper) is treated just like a text or CDATA element
func (parse *Parser) parseString(node *ParseNode) error {
	pt := ParseNode{
		NodeType: NODE_OUTPUT,
		NodeName: "output",
		Data:     parse.current().Literal,
		Children: make([]ParseNode, 0),
		Token:    parse.current(),
		Parent:   node,
	}
	node.Children = append(node.Children, pt)
	return nil
}

func (parse *Parser) parseText(node *ParseNode) error {
	pt := ParseNode{
		NodeType: NODE_OUTPUT,
		NodeName: "output",
		Data:     parse.current().Literal,
		Children: make([]ParseNode, 0),
		Parent:   node,
		Token:    parse.current(),
	}
	node.Children = append(node.Children, pt)
	return nil
}

func (parse *Parser) parseSpecialComment(node *ParseNode) error {
	pt := ParseNode{
		NodeType: NODE_NOP,
		NodeName: "comment",
		Data:     parse.current().Literal,
		Children: make([]ParseNode, 0),
		Parent:   node,
		Token:    parse.current(),
	}
	node.Children = append(node.Children, pt)
	return nil
}

func (parse *Parser) parseSpecialCData(node *ParseNode) error {
	pt := ParseNode{
		NodeType: NODE_OUTPUT,
		NodeName: "cdata",
		Data:     parse.current().Literal,
		Children: make([]ParseNode, 0),
		Parent:   node,
		Token:    parse.current(),
	}
	node.Children = append(node.Children, pt)
	return nil
}

func (parse *Parser) parseSpecialOtherTag(node *ParseNode) error {
	pt := ParseNode{
		NodeType: NODE_SPECIAL_OTHER,
		NodeName: "special",
		Data:     parse.current().Literal,
		Children: make([]ParseNode, 0),
		Parent:   node,
		Token:    parse.current(),
	}
	node.Children = append(node.Children, pt)
	return nil
}

func (parse *Parser) parseCloseTag(node *ParseNode) error {
	pt := ParseNode{
		NodeType: NODE_CLOSE_TAG,
		NodeName: parse.current().Literal,
		Data:     parse.current().Literal,
		Children: make([]ParseNode, 0),
		Parent:   node,
		Token:    parse.current(),
	}
	node.Children = append(node.Children, pt)
	return nil
}
