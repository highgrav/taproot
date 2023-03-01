package jsmltranspiler

import (
	"fmt"
	"highgrav/taproot/v1/languages/jsmlparser"
	"highgrav/taproot/v1/languages/lexer"
	"testing"
)

func TestParse(t *testing.T) {
	const input string = `
<html>
	<head display=true>
	<meta header="foobar" size=123/>
	<title>Hello, world!</title>
	<!DATA ... >
</head>

	<!--
		This is a comment
	-->
	<body @@csp_nonce style="font-size:10em;">
		<h1>🥇🥇 GoldFusion Test 🥇🥇</h1>
		This is a test
		<ul id="myUL">
		<go>
			var total = 0;
			for(var x = 0; x < 10; x++){
				<li>
				<go.val>x</go.val>
				total++;
				</li>
			}
		</go>
		</ul>
	</body>
</html>
<![CDATA[0123456789]]>
<url is-good = "yes"/>
	`
	lex := lexer.New(input)
	toks, err := lex.Lex()
	if err != nil {
		t.Error(err)
	}
	parse := jsmlparser.New(&toks, input)
	err = parse.Parse()
	if err != nil {
		t.Error(err)
	}
	tr := NewWithNode(parse.Tree(), true)
	/*
		fmt.Println(tr.ToString())
	*/

	err = tr.ToJS()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tr.output.String())
}
