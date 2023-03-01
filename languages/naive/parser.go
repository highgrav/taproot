package naive

import (
	"strings"
)

func ParseGoldfusionToJS(gfsrc string) (string, error) {
	res, err := tokenizeHtml(strings.NewReader(gfsrc))
	if err != nil {
		return "", err
	}
	page, err := translateTokensToNodes(res)
	if err != nil {
		return "", err
	}
	return translateNodesToJs(page), nil
}
