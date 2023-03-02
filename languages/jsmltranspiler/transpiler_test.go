package jsmltranspiler

import (
	"fmt"
	"highgrav/taproot/v1/languages/jsmlparser"
	"highgrav/taproot/v1/languages/lexer"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	const input string = `<html lang="en">
	<head display=true>
	<meta header="foobar" size=123/>
	<title>Hello, world!</title>
	<!DATA ... >
</head>

	<!--
		This is a comment
	-->
	<body @@csp_nonce style="font-size:1em;">
		<h1>🥇🥇 GoldFusion Test 🥇🥇</h1>
		This is a test
		<ul id="myUL" style="color:red">
		<go>
			var total = 0;
			for(var x = 0; x < 10; x++){
				<li>
				<go.out>Current value:</go.out> <go.val>x</go.val>
				out.write("OK");
				total++;
				</li>
			}
		</go>
		</ul>
		<p>
			<![CDATA[0123456789]]>
		</p>
		<url is-good="yes"/>
	</body>
</html>`
	lex := lexer.New(input)
	toks, err := lex.Lex()
	if err != nil {
		t.Error(err)
	}

	/*
		for i, v := range toks {
			fmt.Printf("%d: %s\n", i, v.Dump())
		}
	*/

	parse := jsmlparser.New(&toks, input)
	err = parse.Parse()
	if err != nil {
		t.Error(err)
	}
	tr := NewWithNode(parse.Tree(), true)

	err = tr.ToJS()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tr.output.String())
	os.WriteFile("/tmp/test.js", []byte(tr.output.String()), 777)
}
