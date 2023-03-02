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


func TestReadDoc(t *testing.T) {
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
	l := New(input)
	toks, err := l.Lex()
	if err != nil {
		t.Error(err)
	}
	for i, v := range toks {
		fmt.Printf("%d: %s: %s\n", i, v.Type, v.Literal)
	}
}
****************/

func TestReadProblemDoc(t *testing.T) {
	var input string = `<html>
    <head>
        <title>SOME TITLE</title>
    </head>
    <body>
    <div style="color:red;font-style:italic" >
    <go>
        if(db == undefined){
            <go.out>DB not defined!</go.out>
        } else if (db["dsns"] == undefined) {
            <go.out>No dsns() defined!</go.out>
        } else {
            var dsnlist = db.dsns();
            <go.out>Type:</go.out> <go.val>typeof(dsnlist.results.dsns)</go.val>
            <br/>
            <go.val>JSON.stringify(dsnlist.results.dsns)</go.val>
            <br/>
            <go.out>Length: </go.out><go.val>dsnlist.results.dsns.length</go.val>
            <br/>
           <ul>
            if(dsnlist.ok){
                for(var x = 0; x < dsnlist.results.dsns.length; x++){
                    <li>
                    <go.out>DSN </go.out><go.val>x</go.val><go.out>:</go.out>
					<span style="color:red;">
                        <i> <go.val>dsnlist.results.dsns[x]</go.val></i>
                    </span>
                    </li>
                }
            }
            </ul>
        }
        if(db["query"] == undefined) {
            <go.out>No query() defined!</go.out>
        } else {
            var results = db.query("db", "select * from messages");
            <go.val>JSON.stringify(results)</go.val>
        }
    </go>
    </div>
        <h3>List Example</h3>
        <ul>
			<go>
            var totalItems = 0;
            for(var x = 0; x < 10; x++){
                <li>
                    totalItems++;
                    <go.out>Item number </go.out><go.val>x</go.val>
                </li>
            }
			</go>
        </ul>
    </body>
</html>`
	l := New(input)
	toks, err := l.Lex()
	if err != nil {
		t.Error(err)
	}
	for _, v := range toks {
		fmt.Printf("(%s)%s|", v.Type, v.Literal)
	}
}
