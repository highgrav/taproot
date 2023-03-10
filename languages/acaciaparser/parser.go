package acaciaparser

import (
	"fmt"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/languages/lexer"
	"highgrav/taproot/v1/languages/token"
)

type AcaciaNodeType string

const (
	NODE_START     AcaciaNodeType = "start"
	NODE_POLICY    AcaciaNodeType = "policy"
	NODE_MANIFEST  AcaciaNodeType = "manifest"
	NODE_ATTRIBUTE AcaciaNodeType = "attribute"
	NODE_PATHS     AcaciaNodeType = "paths"
	NODE_PATH      AcaciaNodeType = "path"
	NODE_EFFECTS   AcaciaNodeType = "effects"
	NODE_RIGHTS    AcaciaNodeType = "rights"
	NODE_RIGHT     AcaciaNodeType = "right"
	NODE_DENY      AcaciaNodeType = "deny"
	NODE_REDIRECT  AcaciaNodeType = "redirect"
	NODE_LOGS      AcaciaNodeType = "logs"
	NODE_LOG_GROUP AcaciaNodeType = "loggroup"
	NODE_LOG       AcaciaNodeType = "log"
	NODE_MATCHES   AcaciaNodeType = "matches"
	NODE_MATCH     AcaciaNodeType = "match"
	NODE_ERROR     AcaciaNodeType = "error"
	NODE_EOF       AcaciaNodeType = "eof"
)

type AcaciaParseNode struct {
	NodeType AcaciaNodeType
	NodeName string
	Data     string
	Code     int
	Children []AcaciaParseNode
	Parent   *AcaciaNodeType
	Token    token.Token
}

type AcaciaParser struct {
	script string
	tokens *[]token.Token
	nodes  []AcaciaParseNode
}

func New(script string) (*AcaciaParser, error) {
	ap := &AcaciaParser{
		script: script,
		nodes:  make([]AcaciaParseNode, 0),
	}
	l := lexer.New(script)
	toks, err := l.Lex()
	if err != nil {
		return nil, err
	}
	ap.tokens = &toks
	return ap, nil
}

func (p *AcaciaParser) current() *AcaciaParseNode {
	if p.nodes == nil || len(p.nodes) == 0 {
		return nil
	}
	return &(p.nodes[len(p.nodes)-1])
}

func (p *AcaciaParser) Parse() (acacia.Policy, error) {
	policy := acacia.Policy{}
	p.nodes = make([]AcaciaParseNode, 0)

	p.nodes = append(p.nodes, AcaciaParseNode{
		NodeType: NODE_START,
		NodeName: "",
		Data:     "",
		Code:     0,
		Children: make([]AcaciaParseNode, 0),
		Parent:   nil,
		Token:    token.Token{},
	})

	for _, n := range *p.tokens {
		fmt.Printf("%s: %s\n", n.Type, n.Literal)
	}

	return policy, nil
}
