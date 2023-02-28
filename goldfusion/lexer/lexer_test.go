package lexer

import (
	"fmt"
	"testing"
)

/**
func TestComment(t *testing.T) {
	input := `<!-- THIS IS A <test> " +
		"-->`
	l := New(input)
	toks := l.readTag()
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}


func TestComplexSelfClosingTag(t *testing.T) {
	input := "<go.prop test=\"foo()\" style=\"foo:bar;\" />"
	l := New(input)
	toks := l.readOpenTag()
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}


func TestSimpleSelfClosingTag(t *testing.T) {
	input := "<go.prop />"
	l := New(input)
	toks := l.readOpenTag()
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}

// Eyeball tests for reading plain text

func TestSimpleOpenTag(t *testing.T) {
	input := "<go>"
	l := New(input)
	toks := l.readOpenTag()
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}

/********
func TestEndTag(t *testing.T) {
	input := `</tag>`
	l := New(input)
	toks := l.readCloseTag()

	if len(toks) != 1 && toks[0].Type != token.TOKEN_CLOSE_TAG && toks[0].Literal != "</tag>" {
		t.Error("incorrect return from readCloseTag()")
	}
}

func TestReadText(t *testing.T) {
	input := `
for(x = 10; x >= 0; x--) {
	http.write(x);
}
	`
	l := New(input)
	toks := l.readText()
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}
****************/

func TestReadDoc(t *testing.T) {
	const input string = `
<html>
	<head></head>
	<!--
		This is a comment
	-->
	<body style="font-size:10em;">
		<h1>Test</h1>
		This is a test
		<ul id="myUL">
		<go>
			var total = 0;
			for(var x = 0; x < 10; x++){
				<go.out><li></go.out>
				<go.val>x</go.val>
				total++;
				<go.out></li></go.out>
			}
		</go>
		</ul>
	</body>
</html>
	`
	l := New(input)
	toks, err := l.Lex()
	if err != nil {
		t.Error(err)
	}
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}
