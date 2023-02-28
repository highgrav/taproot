package naive

import "golang.org/x/net/html"

type pageNodeType string

const (
	typeHTML pageNodeType = "html"
	typeJS   pageNodeType = "js"
)

type pageNode struct {
	NodeType pageNodeType
	Raw      string
	Token    html.Token
	Children []pageNode
}
