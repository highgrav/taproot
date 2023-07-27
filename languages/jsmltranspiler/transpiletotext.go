package jsmltranspiler

import "github.com/highgrav/taproot/languages/jsmlparser"

/* Writes a parse tree back out to something that should look like the original input. */
func (tr *Transpiler) ToString() string {
	return tr.dispatchToString(*tr.tree)
}

func (tf *Transpiler) dispatchToString(node jsmlparser.ParseNode) string {
	ret := ""
	switch node.NodeType {
	case jsmlparser.NODE_ERROR:
	case jsmlparser.NODE_DOCUMENT:
	case jsmlparser.NODE_EOF:
		ret += ""
	case jsmlparser.NODE_NOP:
		if node.NodeName == "comment" && tf.DisplayComments {
			ret += node.Data
		} else {
			ret += ""
		}
	case jsmlparser.NODE_OUTPUT:
		if node.NodeName == "cdata" {
			ret += "<!CDATA[" + node.Data + "]]>"
		} else {
			ret += node.Data
		}
	case jsmlparser.NODE_TAG:
		ret += "<" + node.NodeName
	case jsmlparser.NODE_SELF_CLOSE_TAG:
		ret += "/>"
	case jsmlparser.NODE_CLOSE_TAG:
		if isTagSemantic(node) {
			ret += " /*" + node.NodeName + "*/ "
		} else {
			ret += node.NodeName
		}
	case jsmlparser.NODE_SPECIAL_OTHER:
		ret += node.Data
	case jsmlparser.NODE_ATTRIBUTE:
		ret += " " + node.NodeName
		if len(node.Children) > 0 {
			ret += "="
		}
	case jsmlparser.NODE_STRING:
		ret += node.Data
	case jsmlparser.NODE_ID:
		ret += node.Data
	case jsmlparser.NODE_NUMBER:
		ret += node.Data
	case jsmlparser.NODE_INTERPOLATED_VALUE:
		ret += node.Data
	default:
	}
	for _, n := range node.Children {
		ret += tf.dispatchToString(n)
	}
	if node.NodeType == jsmlparser.NODE_TAG {
		if node.IsSelfClosingTag {
			ret += "/>"
		} else {
			ret += ">"
		}
	}
	return ret
}
