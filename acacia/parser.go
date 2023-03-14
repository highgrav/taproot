package acacia

import (
	"errors"
	"highgrav/taproot/v1/languages/lexer"
	"highgrav/taproot/v1/languages/token"
	"strconv"
	"strings"
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

func NewParser(script string) (*AcaciaParser, error) {
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

func (p *AcaciaParser) Parse() (Policy, error) {
	return readPolicy(p.tokens)
}

func readPolicy(toks *[]token.Token) (Policy, error) {
	policy := Policy{
		Manifest: PolicyManifest{},
		Routes:   nil,
		Rights:   PolicyRights{},
		Logging:  PolicyLogging{},
		Match:    "",
	}
	var i int = 0
	for i < len(*toks) {
		if (*toks)[i].Type == "startopentag" && (*toks)[i].Literal == "<manifest" {
			err := readManifest(&policy, &i, toks)
			if err != nil {
				return Policy{}, err
			}
		} else if (*toks)[i].Type == "startopentag" && (*toks)[i].Literal == "<paths" {
			err := readPaths(&policy, &i, toks)
			if err != nil {
				return Policy{}, err
			}
		} else if (*toks)[i].Type == "startopentag" && (*toks)[i].Literal == "<effects" {
			err := readEffects(&policy, &i, toks)
			if err != nil {
				return Policy{}, err
			}
		} else if (*toks)[i].Type == "startopentag" && (*toks)[i].Literal == "<log" {
			err := readLogs(&policy, &i, toks)
			if err != nil {
				return Policy{}, err
			}
		} else if (*toks)[i].Type == "startopentag" && (*toks)[i].Literal == "<matches" {
			err := readMatches(&policy, &i, toks)
			if err != nil {
				return Policy{}, err
			}
		}
		i++
	}

	return policy, nil
}

func readStringArraysFromElement(i *int, toks *[]token.Token) []string {
	strs := make([]string, 0)
	tok := (*toks)[*i]
	isInText := false
	for tok.Type != "closetag" && tok.Type != "eof" && tok.Type != "error" {
		if isInText && tok.Type == "string" {
			strs = append(strs, tok.Literal)
		} else if tok.Type == "endopentag" {
			isInText = true
		}
		*i++
		tok = (*toks)[*i]
	}
	return strs
}

// quick convenience function to get everything from between sets
func readTextFromElement(i *int, toks *[]token.Token) string {
	sb := strings.Builder{}
	tok := (*toks)[*i]
	isInText := false
	for tok.Type != "closetag" && tok.Type != "eof" && tok.Type != "error" {
		if isInText {
			sb.Write([]byte(tok.Literal))
		} else if tok.Type == "endopentag" {
			isInText = true
		}
		*i++
		tok = (*toks)[*i]
	}
	// don't advance the end of the tag, the loop this is calling from should do that
	return sb.String()
}

func readManifest(p *Policy, i *int, toks *[]token.Token) error {
	tok := (*toks)[*i]
	for tok.Type != "closetag" && tok.Type != "eof" && tok.Type != "error" && tok.Literal != "</manifest>" {
		if tok.Type == "startopentag" && tok.Literal == "<ns" {
			val := readTextFromElement(i, toks)
			p.Manifest.Namespace = val
		} else if tok.Type == "startopentag" && tok.Literal == "<v" {
			val := readTextFromElement(i, toks)
			p.Manifest.Version = val
		} else if tok.Type == "startopentag" && tok.Literal == "<name" {
			val := readTextFromElement(i, toks)
			p.Manifest.Name = val
		} else if tok.Type == "startopentag" && tok.Literal == "<desc" {
			val := readTextFromElement(i, toks)
			p.Manifest.Description = val
		} else if tok.Type == "startopentag" && tok.Literal == "<priority" {
			val := readTextFromElement(i, toks)
			i, err := strconv.Atoi(val)
			if err != nil {
				return errors.New("attempting to read priority from manifest (should be integer): " + err.Error())
			}
			p.Manifest.Priority = i
		} else if tok.Type == "startopentag" && tok.Literal == "<id" {
			val := readTextFromElement(i, toks)
			p.Manifest.ID = val
		}
		// advance
		*i++
		tok = (*toks)[*i]
	}
	if tok.Type == "eof" {
		return errors.New("unexpected eof, manifest section not closed")
	} else if tok.Type == "error" {
		return errors.New("unexpected error: " + tok.Literal)
	}
	return nil
}

func readPaths(p *Policy, i *int, toks *[]token.Token) error {
	tok := (*toks)[*i]
	for tok.Type != "closetag" && tok.Type != "eof" && tok.Type != "error" && tok.Literal != "</paths>" {
		if tok.Type == "startopentag" && tok.Literal == "<path" {
			val := readTextFromElement(i, toks)
			if p.Routes == nil {
				p.Routes = make([]string, 0)
			}
			p.Routes = append(p.Routes, val)
		}
		*i++
		tok = (*toks)[*i]
	}
	if tok.Type == "eof" {
		return errors.New("unexpected eof, paths section not closed")
	} else if tok.Type == "error" {
		return errors.New("unexpected error: " + tok.Literal)
	}
	return nil
}

func readEffects(p *Policy, i *int, toks *[]token.Token) error {
	tok := (*toks)[*i]
	for tok.Type != "closetag" && tok.Type != "eof" && tok.Type != "error" && tok.Literal != "</effects>" {
		if tok.Type == "startopentag" && tok.Literal == "<allow" {
			rights := readStringArraysFromElement(i, toks)

			p.Rights.Allowed = make([]string, len(rights))
			for _, v := range rights {
				p.Rights.Allowed = append(p.Rights.Allowed, v[1:len(v)-1])
			}
		} else if tok.Type == "startopentag" && tok.Literal == "<deny" {
			rights := readStringArraysFromElement(i, toks)
			p.Rights.Denied = make([]string, len(rights))
			for _, v := range rights {
				p.Rights.Denied = append(p.Rights.Denied, v[1:len(v)-1])
			}
		} else if tok.Type == "startopentag" && tok.Literal == "<redirect" {
			redirectTo := readTextFromElement(i, toks)
			if strings.HasPrefix(redirectTo, "\"") {
				p.Rights.Redirect = redirectTo[1 : len(redirectTo)-1]
			} else {
				p.Rights.Redirect = redirectTo
			}
		} else if tok.Type == "startopentag" && tok.Literal == "<return" {
			returnMsg := readTextFromElement(i, toks)
			p.Rights.ReturnMsg = returnMsg
		} else if tok.Type == "startopentag" && tok.Literal == "<returncode" {
			returnCode := readTextFromElement(i, toks)
			rcode, err := strconv.Atoi(returnCode)
			if err != nil {
				return errors.New("attempting to read return code (should be integer): " + err.Error())
			}
			p.Rights.ReturnCode = rcode
		}
		*i++
		tok = (*toks)[*i]
	}
	if tok.Type == "eof" {
		return errors.New("unexpected eof, effects section not closed")
	} else if tok.Type == "error" {
		return errors.New("unexpected error: " + tok.Literal)
	}
	return nil
}

func readLogs(p *Policy, i *int, toks *[]token.Token) error {
	*i++
	return nil
}

func readMatches(p *Policy, i *int, toks *[]token.Token) error {
	tok := (*toks)[*i]
	for tok.Type != "closetag" && tok.Type != "eof" && tok.Type != "error" && tok.Literal != "</paths>" {
		if tok.Type == "startopentag" && tok.Literal == "<match" {
			val := readTextFromElement(i, toks)
			p.Match = val
		}
		*i++
		tok = (*toks)[*i]
	}
	if tok.Type == "eof" {
		return errors.New("unexpected eof, matches section not closed")
	} else if tok.Type == "error" {
		return errors.New("unexpected error: " + tok.Literal)
	}
	return nil
}
