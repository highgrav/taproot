package gfparser

import "fmt"

func DumpNode(node ParseNode, depth int) {
	depth++
	fmt.Printf("%*sType:%s, Name: %s, Literal: %s\n", depth, "", node.NodeType, node.NodeName, node.Data)
	for _, v := range node.Children {
		PrintNode(v, depth)
	}
}

func PrintNode(node ParseNode, depth int) {
	depth++
	fmt.Printf("%s", node.Data)
	for _, v := range node.Children {
		PrintNode(v, depth)
	}
}
