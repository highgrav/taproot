package pages

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

type translatorNodeType string

const (
	typeStart             translatorNodeType = "start"
	typeTextOutput        translatorNodeType = "text"
	typeGoOutput          translatorNodeType = "goout"
	typeStartScript       translatorNodeType = "go"
	typeScript            translatorNodeType = "script"
	typeVariableOutput    translatorNodeType = "varout"
	typeEnd               translatorNodeType = "end"
	typeEndScript         translatorNodeType = "endscript" // matches script
	typeEndVariableOutput translatorNodeType = "endvarout" // matches varout
	typeEndGoOutput       translatorNodeType = "endgoout"  // matches goout
)

type translatorNode struct {
	nodeType translatorNodeType
	text     string
}

type translatorPage struct {
	nodes []translatorNode
}

func (tp *translatorPage) currentNodeType() translatorNodeType {
	if tp.last().nodeType == typeEnd {
		return typeEnd
	}
	if tp.last().nodeType == typeStart {
		return typeStart
	}

	countScript := 0
	countOut := 0
	countVarOut := 0
	for x := len(tp.nodes) - 1; x >= 0; x-- {
		if tp.nodes[x].nodeType == typeEndScript {
			countScript++
		}
		if tp.nodes[x].nodeType == typeEndGoOutput {
			countOut++
		}
		if tp.nodes[x].nodeType == typeEndVariableOutput {
			countVarOut++
		}
		if tp.nodes[x].nodeType == typeStartScript {
			countScript--
		}
		if tp.nodes[x].nodeType == typeGoOutput {
			countOut--
		}
		if tp.nodes[x].nodeType == typeVariableOutput {
			countVarOut--
		}
		if countScript < 0 {
			return typeScript
		}
		if countOut < 0 {
			return typeGoOutput
		}
		if countVarOut < 0 {
			return typeVariableOutput
		}
	}

	return typeStart
}

func (tp *translatorPage) last() *translatorNode {
	return &(tp.nodes[len(tp.nodes)-1])
}

func (tp *translatorPage) addNode(node translatorNode) {
	tp.nodes = append(tp.nodes, node)
}

func (tp *translatorPage) addNewNode(nodeType translatorNodeType, text string) {
	tp.nodes = append(tp.nodes, translatorNode{
		nodeType: nodeType,
		text:     text,
	})
}

func translateNodesToJson(page translatorPage) string {
	var sb strings.Builder

	sb.Write([]byte("{\"data\":[\n"))
	for i, node := range page.nodes {
		sb.Write([]byte(fmt.Sprintf("{\"id\":%d},\"type\":\"%s\",\"text\":\"%s\"},\n", i, node.nodeType, node.text)))
	}
	sb.Write([]byte("\n]}"))
	return sb.String()
}

func translateNodesToJs(page translatorPage) string {
	var sb strings.Builder
	for _, node := range page.nodes {
		if node.nodeType == typeEndScript || node.nodeType == typeEndGoOutput || node.nodeType == typeEndVariableOutput || node.nodeType == typeStart || node.nodeType == typeEnd {
			continue
		}
		if node.nodeType == typeTextOutput {
			sb.Write([]byte("html.write(\"" + escapeTextForWriting(node.text) + "\");\n"))
			continue
		}
		if node.nodeType == typeVariableOutput {
			sb.Write([]byte("html.write(" + node.text + ");\n"))
			continue
		}
		if node.nodeType == typeScript {
			sb.Write([]byte(node.text))
			continue
		}
	}
	return sb.String()
}

func escapeTextForWriting(str string) string {
	var chars = map[string]bool{
		"\"": true,
		"'":  true,
	}

	var escapeStr = ""

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

func translateTokensToNodes(toks []html.Token) (translatorPage, error) {
	page := translatorPage{nodes: make([]translatorNode, 0)}
	page.addNewNode(typeStart, "")

	for _, tok := range toks {

		// We ignore comments
		if tok.Type == html.CommentToken {
			continue
		}

		if tok.Type == html.DoctypeToken {
			translateDocTypeToken(&page, tok)
			continue
		}

		if tok.Type == html.TextToken {
			translateTextToken(&page, tok)
			continue
		}

		if tok.Type == html.StartTagToken {
			translateStartTagToken(&page, tok)
			continue
		}

		if tok.Type == html.EndTagToken {
			translateEndTagToken(&page, tok)
			continue
		}

		if tok.Type == html.SelfClosingTagToken {
			translateSelfClosingTagToken(&page, tok)
			continue
		}

		if tok.Type == html.ErrorToken {
			return page, errors.New(tok.Data)
		}
	}

	page.addNewNode(typeEnd, "")
	return page, nil
}

func translateDocTypeToken(page *translatorPage, tok html.Token) {
	page.addNewNode(typeTextOutput, tok.Data)
}

func translateTextToken(page *translatorPage, tok html.Token) {
	if len(strings.TrimSpace(tok.Data)) == 0 {
		return
	}
	if page.last().nodeType == typeScript {
		page.last().text = page.last().text + tok.Data
	} else if page.last().nodeType == typeVariableOutput {
		page.last().text = page.last().text + tok.Data
	} else if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + tok.Data
	} else {
		switch page.currentNodeType() {
		case typeStart:
			page.addNewNode(typeTextOutput, tok.Data)
			return
		case typeScript:
			page.addNewNode(typeScript, tok.Data)
			return
		case typeVariableOutput:
			page.addNewNode(typeVariableOutput, tok.Data)
			return
		case typeTextOutput:
			page.addNewNode(typeTextOutput, tok.Data)
			return
		default:
			return
		}
	}
}

func translateStartTagToken(page *translatorPage, tok html.Token) {
	if tok.Data == "go" {
		page.addNewNode(typeStartScript, "")
		page.addNewNode(typeScript, "")
		return
	} else if tok.Data == "go.var" {
		page.addNewNode(typeVariableOutput, "")
		return
	} else if tok.Data == "go.out" {
		page.addNewNode(typeGoOutput, "")
		page.addNewNode(typeTextOutput, "")
		return
	}

	// otherwise, this is a standard tag, so just output it
	if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + "<" + tok.Data + ">"
	} else {
		page.addNewNode(typeTextOutput, "<"+tok.Data+">")
	}
}

func translateEndTagToken(page *translatorPage, tok html.Token) {
	if tok.Data == "go" {
		// ignore this tag, everything inside is a script tag
		page.addNewNode(typeEndScript, "")
		return
	} else if tok.Data == "go.var" {
		page.addNewNode(typeEndVariableOutput, "")
		return
	} else if tok.Data == "go.out" {
		page.addNewNode(typeEndGoOutput, "")
		return
	}

	// otherwise, this is a standard tag, so just output it
	if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + "</" + tok.Data + ">"
	} else {
		page.addNewNode(typeTextOutput, "</"+tok.Data+">")
	}
}

func translateSelfClosingTagToken(page *translatorPage, tok html.Token) {
	if tok.Data == "go" || tok.Data == "go.var" || tok.Data == "go.out" {
		// none of these can be self-closing tags, so ignore them
		return
	}
	if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + "<" + tok.Data + "/>"
	} else {
		page.addNewNode(typeTextOutput, "<"+tok.Data+"/>")
	}
}
