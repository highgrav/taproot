package jsmlparser

import (
	"highgrav/taproot/v1/languages/jsmltranspiler"
	"highgrav/taproot/v1/languages/lexer"
	"testing"
)

func TestParse(t *testing.T) {
	const input string = `
<html>
	<head display=true></head>
	<meta header="foobar" size=123/>
	<!DATA ... >
	<!--
		This is a comment
	-->
	<body style="font-size:10em;">
		<h1>🥇🥇 GoldFusion Test 🥇🥇</h1>
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
<![CDATA[<hello> this is mission control]]>
<url is-good = "yes"/>
	`
	lex := lexer.New(input)
	toks, err := lex.Lex()
	if err != nil {
		t.Error(err)
	}
	parse := New(&toks, input)
	err = parse.Parse()
	if err != nil {
		t.Error(err)
	}
	tr := jsmltranspiler.NewWithNode(parse.tree, true)
	tr.ToString()
}
