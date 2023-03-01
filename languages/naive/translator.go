package naive

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

type translatorNodeType string

const (
	TAG_GO         string = "go"
	TAG_GO_OUT     string = "go.out"
	TAG_GO_VAL     string = "go.val"
	TAG_GO_INCLUDE string = "go.include"
	TAG_GO_GET     string = "go.get"
	TAG_GO_OBJ     string = "go.obj"
	TAG_GO_PROP    string = "go.prop"
)

const (
	typeStart          translatorNodeType = "start"
	typeTextOutput     translatorNodeType = "text"
	typeGoOutput       translatorNodeType = "goout"
	typeStartScript    translatorNodeType = "go"
	typeScript         translatorNodeType = "script"
	typeValueOutput    translatorNodeType = "valout"
	typeEnd            translatorNodeType = "end"
	typeEndScript      translatorNodeType = "endscript" // matches script
	typeEndValueOutput translatorNodeType = "endvalout" // matches varout
	typeEndGoOutput    translatorNodeType = "endgoout"  // matches goout
)

type translatorNode struct {
	nodeType translatorNodeType
	text     string
	attrs    map[string]string
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
		if tp.nodes[x].nodeType == typeEndValueOutput {
			countVarOut++
		}
		if tp.nodes[x].nodeType == typeStartScript {
			countScript--
		}
		if tp.nodes[x].nodeType == typeGoOutput {
			countOut--
		}
		if tp.nodes[x].nodeType == typeValueOutput {
			countVarOut--
		}
		if countScript < 0 {
			return typeScript
		}
		if countOut < 0 {
			return typeGoOutput
		}
		if countVarOut < 0 {
			return typeValueOutput
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
		attrs:    make(map[string]string),
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
		if node.nodeType == typeEndScript || node.nodeType == typeEndGoOutput || node.nodeType == typeEndValueOutput || node.nodeType == typeStart || node.nodeType == typeEnd {
			continue
		}
		if node.nodeType == typeTextOutput {
			sb.Write([]byte("http.write(\"" + escapeTextForWriting(node.text) + "\");\n"))
			continue
		}
		if node.nodeType == typeValueOutput {
			sb.Write([]byte("http.write(" + node.text + ");\n"))
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
	} else if page.last().nodeType == typeValueOutput {
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
		case typeValueOutput:
			page.addNewNode(typeValueOutput, tok.Data)
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
	if tok.Data == TAG_GO {
		page.addNewNode(typeStartScript, "")
		page.addNewNode(typeScript, "")
		return
	} else if tok.Data == TAG_GO_VAL {
		page.addNewNode(typeValueOutput, "")
		return
	} else if tok.Data == TAG_GO_OUT {
		page.addNewNode(typeGoOutput, "")
		page.addNewNode(typeTextOutput, "")
		return
	}

	// otherwise, this is a standard tag, so just output it
	if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + outputTagToken(tok)
	} else {
		page.addNewNode(typeTextOutput, outputTagToken(tok))
	}
}

func translateEndTagToken(page *translatorPage, tok html.Token) {
	if tok.Data == TAG_GO {
		// ignore this tag, everything inside is a script tag
		page.addNewNode(typeEndScript, "")
		return
	} else if tok.Data == TAG_GO_VAL {
		page.addNewNode(typeEndValueOutput, "")
		return
	} else if tok.Data == TAG_GO_OUT {
		page.addNewNode(typeEndGoOutput, "")
		return
	}

	// otherwise, this is a standard tag, so just output it
	if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + outputTagToken(tok)
	} else {
		page.addNewNode(typeTextOutput, outputTagToken(tok))
	}
}

func translateSelfClosingTagToken(page *translatorPage, tok html.Token) {
	if tok.Data == TAG_GO || tok.Data == TAG_GO_VAL || tok.Data == TAG_GO_OUT {
		// none of these can be self-closing tags, so ignore them
		return
	}
	if page.last().nodeType == typeTextOutput {
		page.last().text = page.last().text + outputTagToken(tok)
	} else {
		page.addNewNode(typeTextOutput, outputTagToken(tok))
	}
}

func outputTagToken(token html.Token) string {
	var sb strings.Builder
	if token.Type != html.StartTagToken && token.Type != html.EndTagToken && token.Type != html.SelfClosingTagToken {
		return token.Data
	}
	if token.Type == html.EndTagToken {
		sb.Write([]byte("</"))
		sb.Write([]byte(token.Data))
		sb.Write([]byte(">"))
		return sb.String()
	}

	sb.Write([]byte("<"))
	sb.Write([]byte(token.Data))
	for _, attr := range token.Attr {
		sb.Write([]byte(" "))
		sb.Write([]byte(attr.Key))
		sb.Write([]byte("=\""))
		sb.Write([]byte(escapeTextForWriting(attr.Val)))
		sb.Write([]byte("\" "))
	}

	if token.Type == html.SelfClosingTagToken {
		sb.Write([]byte("/>"))
	} else {
		sb.Write([]byte(">"))
	}
	return sb.String()
}
