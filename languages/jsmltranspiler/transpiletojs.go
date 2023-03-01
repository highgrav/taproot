package jsmltranspiler

import (
	"errors"
	"fmt"
	"highgrav/taproot/v1/languages/jsmlparser"
	"strings"
)

func (tr *Transpiler) ToJS() error {
	tr.output = strings.Builder{}
	tr.modes = make([]transpMode, 1)
	tr.modes = append(tr.modes, transpModeHTMLOutput)
	return tr.dispatchToJS(*tr.tree)
}

func (tr *Transpiler) dispatchToJS(node jsmlparser.ParseNode) error {
	switch node.NodeType {
	case jsmlparser.NODE_DOCUMENT:
		for _, n := range node.Children {
			err := tr.dispatchToJS(n)
			if err != nil {
				return err
			}
		}
		return nil
	case jsmlparser.NODE_ERROR:
		return errors.New(node.Data)
	case jsmlparser.NODE_EOF:
		// Nothing to do here
	case jsmlparser.NODE_NOP:
		if node.NodeName == "comment" {
			// dispatch to comment in case we want to keep them
		} else {
			// Otherwise, ignore!
			return nil
		}
	case jsmlparser.NODE_OUTPUT:
		return tr.dispatchOutput(node)
	case jsmlparser.NODE_STRING:
		return tr.dispatchOutput(node)
	case jsmlparser.NODE_TAG:
		return tr.dispatchTag(node)
	case jsmlparser.NODE_SELF_CLOSE_TAG:
		return tr.throwError(node, "unexpected tag close")
	case jsmlparser.NODE_CLOSE_TAG:
		return tr.dispatchCloseTag(node)
	case jsmlparser.NODE_SPECIAL_OTHER:
		return tr.dispatchSpecialOtherTag(node)

	case jsmlparser.NODE_ATTRIBUTE:
		return tr.throwError(node, "unexpected attribute outside tag")

	case jsmlparser.NODE_ID:
		return tr.throwError(node, "unexpected id outside tag")

	case jsmlparser.NODE_NUMBER:
		return tr.throwError(node, "unexpected number outside tag")

	case jsmlparser.NODE_INTERPOLATED_VALUE:
		return tr.dispatchInterpolatedValue(node)
	default:
		// ???
	}
	return nil
}

func (tr *Transpiler) dispatchOutput(node jsmlparser.ParseNode) error {
	if tr.mode() == transpModeDirectOutput {
		// We're outputting code, so just dump it directly to the output
		tr.output.Write([]byte(node.Data))
	} else {
		// We're writing HTML, so wrap each line in out.write() statements
		res := "out.write(\"" + escapeTextForWriting(node.Data) + "\");\n"
		tr.output.Write([]byte(res))
	}
	return nil
}

func (tr *Transpiler) dispatchInterpolatedValue(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchInterpolatedValue")
	return nil
}

func (tr *Transpiler) dispatchSpecialOtherTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchSpecialOtherTag")
	return nil
}

// TODO -- looks good
func (tr *Transpiler) dispatchCloseTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchCloseTag")
	if isTagSemantic(node) {
		return tr.dispatchSemanticCloseTag(node)
	} else {
		return tr.dispatchNonSemanticCloseTag(node)
	}
}

func (tr *Transpiler) dispatchSemanticCloseTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchSemanticCloseTag")

	return nil
}

// TODO -- looks good
func (tr *Transpiler) dispatchNonSemanticCloseTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchNonSemanticCloseTag")
	ret := "out.write(\"" + escapeTextForWriting(node.Data) + "\");\n"
	tr.output.Write([]byte(ret))
	return nil
}

// TODO -- looks good
func (tr *Transpiler) dispatchTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchTag")
	if isTagSemantic(node) {
		return tr.dispatchSemanticTag(node)
	} else {
		return tr.dispatchNonSemanticTag(node)
	}
}

func (tr *Transpiler) dispatchSemanticTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchSemanticTag")
	if node.IsSelfClosingTag {
		// handle attributes
		attrs, body := extractNodes(node.Children, []jsmlparser.ParserNodeType{jsmlparser.NODE_ATTRIBUTE})
		for _, attr := range attrs {
			fmt.Println(attr.Data)
		}

		if len(body) != 0 {
			return tr.throwError(node, "found non-attribute nodes on a self-closing tag")
		}

		return nil
	} else {
		attrs, body := extractNodes(node.Children, []jsmlparser.ParserNodeType{jsmlparser.NODE_ATTRIBUTE})
		for _, attr := range attrs {
			fmt.Println(attr.Data)
		}

		for _, item := range body {
			fmt.Println(item.NodeType)
		}
		return nil
	}
}

// TODO -- Looks good
func (tr *Transpiler) dispatchNonSemanticTag(node jsmlparser.ParseNode) error {
	fmt.Println("dispatchNonSemanticTag")
	ret := ""
	ret += "<" + node.Data

	attrs, body := extractNodes(node.Children, []jsmlparser.ParserNodeType{jsmlparser.NODE_ATTRIBUTE})

	for _, attr := range attrs {
		if len(attr.Children) == 0 {
			// TODO -- handle special variables (@-vars)
			ret += " " + attr.NodeName + " "
		} else {
			ret += " " + attr.NodeName + "=" + attr.Children[0].Data + " "
		}
	}

	if node.IsSelfClosingTag {
		ret += "/>"
		tr.output.Write([]byte("out.write(\"" + escapeTextForWriting(ret) + "\");\n"))
	} else {
		ret += ">"
		tr.output.Write([]byte("out.write(\"" + escapeTextForWriting(ret) + "\");\n"))

		for _, itm := range body {
			err := tr.dispatchToJS(itm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/////////////////////////////////////////////
// Semantic tag handling
/////////////////////////////////////////////

// <go>
func (tr *Transpiler) dispatchGoOpenTag(node jsmlparser.ParseNode) error {
	return nil
}

func (tr *Transpiler) dispatchGoCloseTag(node jsmlparser.ParseNode) error {
	return nil
}

// <go.out>
func (tr *Transpiler) dispatchGoOutOpenTag(node jsmlparser.ParseNode) error {
	tr.modes = append(tr.modes, transpModeHTMLOutput)
	return nil
}

func (tr *Transpiler) dispatchGoOutCloseTag(node jsmlparser.ParseNode) error {
	if tr.modes[len(tr.modes)-1] != transpModeHTMLOutput {
		return errors.New("attempted to close a <go.out/> tag while not processing output")
	}
	tr.modes = tr.modes[:len(tr.modes)-1]
	return nil
}

// <go.val>
func (tr *Transpiler) dispatchGoValOpenTag(node jsmlparser.ParseNode) error {
	return nil
}

func (tr *Transpiler) dispatchGoValCloseTag(node jsmlparser.ParseNode) error {
	return nil
}

// <go.query>
func (tr *Transpiler) dispatchGoQueryOpenTag(node jsmlparser.ParseNode) error {
	return nil
}

func (tr *Transpiler) dispatchGoQueryCloseTag(node jsmlparser.ParseNode) error {
	return nil
}

// <go.loop>
func (tr *Transpiler) dispatchGoLoopOpenTag(node jsmlparser.ParseNode) error {
	return nil
}

func (tr *Transpiler) dispatchGoLoopCloseTag(node jsmlparser.ParseNode) error {
	return nil
}

// <go.get>
func (tr *Transpiler) dispatchGoGetOpenTag(node jsmlparser.ParseNode) error {
	return tr.throwError(node, "NOT IMPLEMENTED")
}

func (tr *Transpiler) dispatchGoGetCloseTag(node jsmlparser.ParseNode) error {
	return tr.throwError(node, "NOT IMPLEMENTED")
}

// <go.CLASS>
func (tr *Transpiler) dispatchGoClassOpenTag(node jsmlparser.ParseNode) error {
	return tr.throwError(node, "NOT IMPLEMENTED")
}

func (tr *Transpiler) dispatchGoClassCloseTag(node jsmlparser.ParseNode) error {
	return tr.throwError(node, "NOT IMPLEMENTED")
}
