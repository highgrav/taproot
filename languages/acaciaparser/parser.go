package acaciaparser

import (
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/languages/lexer"
	"highgrav/taproot/v1/languages/token"
)

type AcaciaNodeType string

const (
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
)

type AcaciaParseNode struct {
	NodeType AcaciaNodeType
	NodeName string
	Data     string
	Code     int
	Children []AcaciaNodeType
	Parent   *AcaciaNodeType
	Token    token.Token
}

type AcaciaParser struct {
	script string
	tokens *[]token.Token
}

func New(script string) (*AcaciaParser, error) {
	ap := &AcaciaParser{
		script: script,
	}
	l := lexer.New(script)
	toks, err := l.Lex()
	if err != nil {
		return nil, err
	}
	ap.tokens = &toks
	return ap, nil
}

func (p *AcaciaParser) Parse() (acacia.Policy, error) {
	policy := acacia.Policy{}
	return policy, nil
}
