package pages

type parseNodeType string

const (
	nodeTypeStart          parseNodeType = "start"
	nodeTypeTextOutput     parseNodeType = "text"
	nodeTypeScript         parseNodeType = "script"
	nodeTypeVariableOutput parseNodeType = "varout"
	nodeTypeEnd            parseNodeType = "end"
	nodeTypeNop            parseNodeType = "nop"
)

type parseTree struct {
}

type parseNode struct {
}
