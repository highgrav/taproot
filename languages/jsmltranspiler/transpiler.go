package jsmltranspiler

import (
	"errors"
	"fmt"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/languages/jsmlparser"
	"highgrav/taproot/v1/languages/lexer"
	"strings"
)

type transpMode string

const (
	transpModeUnknown            transpMode = "unknown"
	transpModeHTMLOutput         transpMode = "html"   // HTML (wrapped in http.write...)
	transpModeDirectOutput       transpMode = "direct" // JS, don't wrap
	transpModeInterpolatedOutput transpMode = "interp"
)

type Transpiler struct {
	tree            *jsmlparser.ParseNode
	output          strings.Builder
	imports         map[string]Transpiler
	script          string
	DisplayComments bool
	// hidden state
	modes []transpMode
}

func NewAndTranspile(script string, displayComments bool) (Transpiler, error) {
	t := Transpiler{
		DisplayComments: displayComments,
		tree:            nil,
		imports:         make(map[string]Transpiler),
		script:          script,
	}

	lex := lexer.New(script)
	toks, err := lex.Lex()
	if err != nil {
		return t, err
	}

	parse := jsmlparser.New(&toks, script)
	err = parse.Parse()
	if err != nil {
		return t, err
	}

	t.tree = parse.Tree()

	return t, nil
}

func NewWithNode(node *jsmlparser.ParseNode, displayComments bool) Transpiler {
	return Transpiler{
		DisplayComments: displayComments,
		tree:            node,
		imports:         make(map[string]Transpiler),
	}
}

func (tr *Transpiler) Builder() *strings.Builder {
	return &tr.output
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
	ignoreRunes := []rune{'\r', '\t'}
	escapeFromRunes := []rune{'\n', '\'', '"'}
	escapeToStrings := []string{"\\n", "\\'", "\\\""}
	return escapeText(str, ignoreRunes, escapeFromRunes, escapeToStrings)
}

func escapeText(str string, ignoreRunes []rune, escapeFromRunes []rune, escapeToStrings []string) string {

	rs := common.NewRuneString(str)
	sb := common.NewRuneStringBuilder()

	for x := int32(0); x < rs.Length; x++ {
		c := rs.Get(x)
		isTouched := false
		for _, v := range ignoreRunes {
			if c == v {
				isTouched = true
			}
		}
		if !isTouched {
			for i, v := range escapeFromRunes {
				if c == v {
					sb.WriteString(escapeToStrings[i])
					isTouched = true
				}
			}
		}
		if !isTouched {
			sb.WriteRune(c)
		}
	}

	return sb.String()
}

func isTagSemantic(node jsmlparser.ParseNode) bool {
	if node.NodeType != jsmlparser.NODE_TAG && node.NodeType != jsmlparser.NODE_CLOSE_TAG {
		return false
	}
	// TODO -- at least two of these conditions are unnecessary -- review jsmlparser
	if node.NodeName == "go" || strings.HasPrefix(node.NodeName, "go.") || node.NodeName == "<go>" || strings.HasPrefix(node.NodeName, "<go.") || node.NodeName == "</go>" || strings.HasPrefix(node.NodeName, "</go.") {
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

	if len(extracted)+len(remainder) != len(children) {
		panic(fmt.Sprintf("%d + %d != %d", len(extracted), len(remainder), len(children)))
	}

	return extracted, remainder
}
