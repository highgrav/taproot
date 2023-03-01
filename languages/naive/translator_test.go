package naive

import (
	"fmt"
	"strings"
	"testing"
)

func TestBasicTranslation(t *testing.T) {
	htmlPage := `
<go>var results = db.query("select * from users");</go>
<html>
    <head>
        <title>SOME TITLE</title>
    </head>
    <body>
        <h3>List Example</h3>
        <ul>
			<go>
            var totalItems = 0;
            for(var x = 0; x < 10; x++){
                log.Info("Loop " + x);
                <go.out>
					<li>
                        <go>totalItems++;</go>
                        Item number <go.var>x</go.var>
					</li>
                </go.out>
            }
			</go>
        </ul>
    </body>
</html>
`
	res, err := tokenizeHtml(strings.NewReader(htmlPage))
	if err != nil {
		t.Error(err.Error())
	}
	page, err := translateTokensToNodes(res)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Printf("Token count: %d\n", len(page.nodes))
	for _, v := range page.nodes {
		fmt.Printf("%s: %s\n\n", v.nodeType, v.text)
	}

	fmt.Println(translateNodesToJs(page))

	fmt.Println(translateNodesToJson(page))
}
