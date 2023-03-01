package naive

import (
	"fmt"
	"strings"
	"testing"
)

func TestBasicTokenizerRaw(t *testing.T) {
	htmlPage := `
var results = db.query("select * from users");
<html>
    <head>
        <title><go.out>Homepage for <go.var>user.name</go.var></go.out></title>
    </head>
    <body>
        <h3>List Example</h3>
        <ul>
            var totalItems = 0;
            for(var x = 0; x < 10; x++){
                log.Info("Loop " + x);
                <li><go.out>
                        <go>totalItems++;</go>
                        Item number <go.var>x</go.var>
                </go.out></li>
            }
        </ul>
    </body>
</html>
`
	res, err := tokenizeHtml(strings.NewReader(htmlPage))
	if err != nil {
		t.Error(err.Error())
	}
	for i, v := range res {
		attrs := ""
		for _, a := range v.Attr {
			attrs = attrs + " " + a.Key + "=" + a.Val
		}
		fmt.Printf("%d: %s: %d[%s]\n	(%s)\n", i, v.Type, len(v.Attr), attrs, v.Data)
	}
}
