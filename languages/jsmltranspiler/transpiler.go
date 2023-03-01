package jsmltranspiler

import (
	"errors"
	"fmt"
	"highgrav/taproot/v1/common"
	"strings"
)

type transpMode string

const (
	transpModeUnknown      transpMode = "unknown"
	transpModeHTMLOutput   transpMode = "html"   // HTML (wrapped in http.write...)
	transpModeDirectOutput transpMode = "direct" // JS, don't wrap
)

type Transpiler struct {
	tree            *jsmlparser.ParseNode
	output          strings.Builder
	imports         map[string]Transpiler
	script          string
	writerPrefix    string
	DisplayComments bool
	// hidden state
	modes []transpMode
}

func NewWithNode(node *jsmlparser.ParseNode, displayComments bool) Transpiler {
	return Transpiler{
		tree:            node,
		DisplayComments: displayComments,
	}
}

func (tr *Transpiler) throwError(node jsmlparser.ParseNode, msg string) error {
	message := msg
	tok := node.Token
	rs := common.NewRuneString(tr.script)
	line, linepos := rs.GetLineAndPos(int32(tok.CharPos))
	if msg == "" {
		message = tok.Message
	}
	return errors.New(fmt.Sprintf("error '%s' at line %d, pos %d, (token type: %s, literal: '%s')", message, line, linepos, tok.Type, tok.Literal))
}

func (tr *Transpiler) mode() transpMode {
	if len(tr.modes) == 0 {
		return transpModeUnknown
	}
	return tr.modes[len(tr.modes)-1]
}

func escapeTextForWriting(str string) string {
	var chars = map[string]bool{
		"\"": true,
		"'":  true,
	}

	var escapeStr = ""

	str = strings.Replace(str, "\r", " ", -1)
	str = strings.Replace(str, "\n", " ", -1)
	for i := 0; i < len(str); i++ {
		var char = string(str[i])

		if chars[char] == true {
			escapeStr += "\\" + char
		} else {
			escapeStr += char
		}
	}

	return escapeStr
}

func isTagSemantic(node jsmlparser.ParseNode) bool {
	if node.NodeType != jsmlparser.NODE_TAG && node.NodeType != jsmlparser.NODE_CLOSE_TAG {
		return false
	}
	if node.NodeName == "<go>" || strings.HasPrefix(node.NodeName, "<go.") || node.NodeName == "</go>" || strings.HasPrefix(node.NodeName, "</go.") {
		return true
	}
	return false
}

// Goes through a node's children, pulls out specified ParserNodeType nodes, and returns them in two slices
func extractNodes(children []jsmlparser.ParseNode, desiredTypes []jsmlparser.ParserNodeType) (extracted []jsmlparser.ParseNode, remainder []jsmlparser.ParseNode) {
	extracted = make([]jsmlparser.ParseNode, 0)
	remainder = make([]jsmlparser.ParseNode, 0)
	for _, v := range children {
		matched := false
		for _, k := range desiredTypes {
			if v.NodeType == k {
				extracted = append(extracted, v)
				matched = true
				break
			}
		}
		if !matched {
			remainder = append(remainder, v)
		}
	}
	return extracted, remainder
}
