package jsmltranspiler

import (
	"fmt"
	"github.com/highgrav/taproot/v1/languages/jsmlparser"
	"github.com/highgrav/taproot/v1/languages/lexer"
	"os"
	"testing"
)

type testScriptAccessor struct{}

func (sa testScriptAccessor) GetJSScriptByID(id string) (string, error) {
	return "<h1>THIS IS AN INCLUDE TEST</h1>", nil
}

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
	<go.include src="foo"/>
</html>`
	lex := lexer.New(input)
	toks, err := lex.Lex()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Parsing...")
	parse := jsmlparser.New(&toks, input)
	err = parse.Parse()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Transpiling...")
	sa := testScriptAccessor{}
	tr := NewWithNode(sa, parse.Tree(), true)

	err = tr.ToJS()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tr.output.String())
	os.WriteFile("/tmp/test.js", []byte(tr.output.String()), 777)
}
