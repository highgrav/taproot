package jsmltranspiler

import (
	"errors"
	"fmt"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/languages/jsmlparser"
	"github.com/highgrav/taproot/v1/languages/lexer"
	"strings"
)

type IScriptAccessor interface {
	GetJSScriptByID(id string) (string, error)
	GetJSMLScriptByID(id string) (string, error)
}

type transpMode string

const (
	transpModeUnknown            transpMode = "unknown"
	transpModeHTMLOutput         transpMode = "html"   // HTML (wrapped in http.write...)
	transpModeDirectOutput       transpMode = "direct" // JS, don't wrap
	transpModeInterpolatedOutput transpMode = "interp"
)

type Transpiler struct {
	ID              string
	scriptAccessor  IScriptAccessor
	tree            *jsmlparser.ParseNode
	output          strings.Builder
	imports         map[string]string
	script          string
	DisplayComments bool
	IVars           int
	// hidden state
	modes []transpMode
}

func (t *Transpiler) GetImports() []string {
	re := make([]string, 0)
	for k, _ := range t.imports {
		re = append(re, k)
	}
	return re
}

func NewAndTranspile(scriptId string, accessor IScriptAccessor, script string, displayComments bool) (Transpiler, error) {
	t := Transpiler{
		ID:              scriptId,
		DisplayComments: displayComments,
		tree:            nil,
		imports:         make(map[string]string),
		script:          script,
		scriptAccessor:  accessor,
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
	err = t.ToJS()
	if err != nil {
		return t, err
	}
	return t, nil
}

func NewWithNode(accessor IScriptAccessor, node *jsmlparser.ParseNode, displayComments bool) Transpiler {
	return Transpiler{
		DisplayComments: displayComments,
		tree:            node,
		imports:         make(map[string]string),
		scriptAccessor:  accessor,
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

func (tr *Transpiler) getInclude(id string, node jsmlparser.ParseNode) (string, error) {
	for compiledScript, ok := tr.imports[id]; ok; {
		return compiledScript, nil
	}

	// We probably have a delimited string coming in, so strip first and last chars
	// TODO -- adding a function to convert a "safe" string to a "native" string might be a good idea
	if id[0] == '"' || id[0] == '\'' {
		id = id[1 : len(id)-1]
	}
	script, err := tr.scriptAccessor.GetJSMLScriptByID(id)
	if err != nil {
		newErr := tr.throwError(node, "could not access included script '"+id+"'")
		newErr = errors.Join(newErr, err)
		return "", newErr
	}
	compiledTr, err := NewAndTranspile(id, tr.scriptAccessor, script, tr.DisplayComments)
	if err != nil {
		newErr := tr.throwError(node, "failed to compile included script '"+id+"'")
		newErr = errors.Join(newErr, err)
		return "", newErr
	}
	res := compiledTr.Builder().String()
	tr.imports[id] = res
	return res, nil
}
