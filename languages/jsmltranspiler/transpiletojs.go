package jsmltranspiler

import (
	"errors"
	"github.com/highgrav/taproot/v1/languages/jsmlparser"
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
		panic("!!!!")
		//		return tr.dispatchInterpolatedValue(node)
	default:
		// ???
	}
	return nil
}

func (tr *Transpiler) dispatchOutput(node jsmlparser.ParseNode) error {
	if tr.mode() == transpModeDirectOutput {
		// We're outputting code, so just dump it directly to the output
		tr.output.Write([]byte(node.Data))
	} else if tr.mode() == transpModeHTMLOutput {
		// We're writing HTML, so wrap each line in out.write() statements
		res := "out.write(\"" + escapeTextForWriting(node.Data) + "\");\n"
		tr.output.Write([]byte(res))
	} else if tr.mode() == transpModeInterpolatedOutput {
		res := "out.write(" + node.Data + ");\n"
		tr.output.Write([]byte(res))
	}
	return nil
}

func (tr *Transpiler) dispatchInterpolatedValue(node jsmlparser.ParseNode) error {

	return nil
}

func (tr *Transpiler) dispatchSpecialOtherTag(node jsmlparser.ParseNode) error {

	return nil
}

// TODO -- looks good
func (tr *Transpiler) dispatchCloseTag(node jsmlparser.ParseNode) error {
	if isTagSemantic(node) {
		return tr.dispatchSemanticCloseTag(node)
	} else {
		return tr.dispatchNonSemanticCloseTag(node)
	}
}

func (tr *Transpiler) dispatchSemanticCloseTag(node jsmlparser.ParseNode) error {
	if node.NodeType == jsmlparser.NODE_CLOSE_TAG && node.NodeName == "</go>" {
		return tr.dispatchGoCloseTag(node)
	} else if node.NodeType == jsmlparser.NODE_CLOSE_TAG && node.NodeName == "</go.out>" {
		return tr.dispatchGoOutCloseTag(node)
	} else if node.NodeType == jsmlparser.NODE_CLOSE_TAG && node.NodeName == "</go.val>" {
		return tr.dispatchGoValCloseTag(node)
	}
	switch node.NodeName {
	case "</go.include>":
		return tr.dispatchGoIncludeCloseTag(node)
	default:
		return tr.throwError(node, "unknown close tag node type "+node.NodeName)
	}

	return nil
}

// TODO -- looks good
func (tr *Transpiler) dispatchNonSemanticCloseTag(node jsmlparser.ParseNode) error {
	ret := "out.write(\"" + escapeTextForWriting(node.Data) + "\");\n"
	tr.output.Write([]byte(ret))
	return nil
}

// TODO -- looks good
func (tr *Transpiler) dispatchTag(node jsmlparser.ParseNode) error {
	isTagSem := isTagSemantic(node)
	if isTagSem {
		return tr.dispatchSemanticTag(node)
	} else {
		return tr.dispatchNonSemanticTag(node)
	}
}

func (tr *Transpiler) dispatchSemanticTag(node jsmlparser.ParseNode) error {
	if node.NodeType != jsmlparser.NODE_TAG {
		return tr.throwError(node, "tried to dispatch semantic tag for wrong tag type")
	}

	// handle <go/> and <go.out/> as special cases
	if node.NodeType == jsmlparser.NODE_TAG && node.NodeName == "go" {
		return tr.dispatchGoOpenTag(node)
	} else if node.NodeType == jsmlparser.NODE_TAG && node.NodeName == "go.out" {
		return tr.dispatchGoOutOpenTag(node)
	} else if node.NodeType == jsmlparser.NODE_TAG && node.NodeName == "go.val" {
		return tr.dispatchGoValOpenTag(node)
	}

	switch node.NodeName {
	case "go.include":
		return tr.dispatchGoIncludeOpenTag(node)
	default:
		return tr.throwError(node, "unknown tag node type "+node.NodeName)
	}
}

// TODO -- Looks good
func (tr *Transpiler) dispatchNonSemanticTag(node jsmlparser.ParseNode) error {

	rets := make([]string, 0)
	interps := make([]int, 0)
	ret := " "
	ret += "<" + node.Data

	attrs, body := extractNodes(node.Children, []jsmlparser.ParserNodeType{jsmlparser.NODE_ATTRIBUTE})

	for _, attr := range attrs {
		if len(attr.Children) == 0 {
			// TODO -- if we handle @-vars here, we need a way to not escape them later.
			if strings.HasPrefix(attr.NodeName, "@") {
				rets = append(rets, ret)
				// mark that the next value is an interpval
				interps = append(interps, len(rets))
				rets = append(rets, attr.NodeName[1:])
				ret = " "
				tr.IVars = tr.IVars + 1
			} else {
				ret += " " + attr.NodeName + " "
			}
		} else {
			ret += " " + attr.NodeName + "="
			if strings.HasPrefix(attr.Children[0].Data, "@") {
				rets = append(rets, ret)
				// mark that the next value is an interpval
				interps = append(interps, len(rets))
				rets = append(rets, attr.Children[0].Data[1:])
				ret = " "
				tr.IVars = tr.IVars + 1
			} else {
				ret += attr.Children[0].Data + " "
			}
		}
	}

	if node.IsSelfClosingTag {
		ret += "/>"
		rets = append(rets, ret)
		for i, ret := range rets {
			isInterp := false
			for _, v := range interps {
				if v == i {
					isInterp = true
				}
			}
			if !isInterp {
				tr.output.Write([]byte("out.write(\"" + escapeTextForWriting(ret) + "\");\n"))
			} else {
				tr.output.Write([]byte("out.write(\"\\\"\");\n"))
				tr.output.Write([]byte("out.write(" + ret + ");\n"))
				tr.output.Write([]byte("out.write(\"\\\"\");\n"))
			}
		}
	} else {
		ret += ">"
		rets = append(rets, ret)
		for i, ret := range rets {
			isInterp := false
			for _, v := range interps {
				if v == i {
					isInterp = true
				}
			}
			if !isInterp {
				tr.output.Write([]byte("out.write(\"" + escapeTextForWriting(ret) + "\");\n"))
			} else {
				tr.output.Write([]byte("out.write(\"\\\"\");\n"))
				tr.output.Write([]byte("out.write(" + ret + ");\n"))
				tr.output.Write([]byte("out.write(\"\\\"\");\n"))
			}
		}

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
	tr.modes = append(tr.modes, transpModeDirectOutput)
	return nil
}

func (tr *Transpiler) dispatchGoCloseTag(node jsmlparser.ParseNode) error {
	if tr.modes[len(tr.modes)-1] != transpModeDirectOutput {
		return errors.New("attempted to close a <go/> tag while not writing raw output")
	}
	tr.modes = tr.modes[:len(tr.modes)-1]
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
	tr.modes = append(tr.modes, transpModeDirectOutput)
	tr.output.Write([]byte("out.write("))
	for _, c := range node.Children {
		err := tr.dispatchToJS(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tr *Transpiler) dispatchGoValCloseTag(node jsmlparser.ParseNode) error {
	tr.output.Write([]byte(");\n"))
	if tr.modes[len(tr.modes)-1] != transpModeDirectOutput {
		return errors.New("attempted to close a <go.val/> tag while not processing output")
	}
	tr.modes = tr.modes[:len(tr.modes)-1]
	return nil
}

// <go.include src="..." />
// TODO --
func (tr *Transpiler) dispatchGoIncludeOpenTag(node jsmlparser.ParseNode) error {
	src := ""
	params, _ := extractNodes(node.Children, []jsmlparser.ParserNodeType{jsmlparser.NODE_ATTRIBUTE})

	for _, p := range params {
		if strings.ToLower(p.NodeName) == "src" {
			for _, v := range p.Children {
				if v.NodeType == jsmlparser.NODE_STRING {
					src = v.Data
				}
			}
		}
	}

	if src == "" {
		return tr.throwError(node, "missing source ID for included script")
	}
	res, err := tr.getInclude(src, node)
	if err != nil {
		return err
	}
	tr.output.Write([]byte(res))
	return nil
}

func (tr *Transpiler) dispatchGoIncludeCloseTag(node jsmlparser.ParseNode) error {
	// doesn't have any effect
	return nil
}

///////////////////////////////////////////////////////////////

// <go.query>
func (tr *Transpiler) dispatchGoQueryOpenTag(node jsmlparser.ParseNode) error {
	return errors.New("NOT IMPLEMENTED")
}

func (tr *Transpiler) dispatchGoQueryCloseTag(node jsmlparser.ParseNode) error {
	return errors.New("NOT IMPLEMENTED")
}

// <go.loop>
func (tr *Transpiler) dispatchGoLoopOpenTag(node jsmlparser.ParseNode) error {
	return errors.New("NOT IMPLEMENTED")
}

func (tr *Transpiler) dispatchGoLoopCloseTag(node jsmlparser.ParseNode) error {
	return errors.New("NOT IMPLEMENTED")
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
