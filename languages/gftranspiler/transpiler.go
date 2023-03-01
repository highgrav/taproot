package gftranspiler

import "highgrav/taproot/v1/languages/gfparser"

type Transpiler struct {
	tree *gfparser.ParseNode
}

func New(node *gfparser.ParseNode) Transpiler {
	return Transpiler{
		tree: node,
	}
}

func (tr *Transpiler) ToString() string {
	return tr.dispatch(*tr.tree)
}

func (tf *Transpiler) dispatch(node gfparser.ParseNode) string {
	ret := ""
	switch node.NodeType {
	case gfparser.NODE_ERROR:
	case gfparser.NODE_DOCUMENT:
	case gfparser.NODE_NOP:
	case gfparser.NODE_EOF:
		ret += ""
	case gfparser.NODE_OUTPUT:
		ret += node.Data
	case gfparser.NODE_TAG:
		ret += "<" + node.NodeName
	case gfparser.NODE_SELF_CLOSE_TAG:
		ret += "/>"
	case gfparser.NODE_CLOSE_TAG:
		ret += node.Data
	case gfparser.NODE_SPECIAL_OTHER:
		ret += node.Data
	case gfparser.NODE_ATTRIBUTE:
		ret += " " + node.NodeName
		if len(node.Children) > 0 {
			ret += "="
		}
	case gfparser.NODE_STRING:
		ret += node.Data
	case gfparser.NODE_ID:
		ret += node.Data
	case gfparser.NODE_NUMBER:
		ret += node.Data
	case gfparser.NODE_INTERPOLATED_VALUE:
		ret += node.Data
	default:
	}
	for _, n := range node.Children {
		ret += tf.dispatch(n)
	}
	if node.NodeType == gfparser.NODE_TAG {
		if node.IsSelfClosingTag {
			ret += "/>"
		} else {
			ret += ">"
		}
	}
	return ret
}
