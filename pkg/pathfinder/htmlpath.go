package pathfinder

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
)

// AntchfxHTMLFinder represents implementation of HTMLPath from https://github.com/antchfx/htmlquery
type AntchfxHTMLFinder struct{}

func NewAntchfxHTMLFinder() AntchfxHTMLFinder {
	return AntchfxHTMLFinder{}
}

func (a AntchfxHTMLFinder) Find(expr string, b []byte) (any, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	nodes, err := htmlquery.QueryAll(doc, expr)
	if len(nodes) >= 1 {
		return any(strings.TrimSpace(nodes[0].FirstChild.Data)), nil
	}

	return nil, fmt.Errorf("could not find %s in given HTML bytes", expr)
}
