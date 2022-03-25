package pathfinder

import (
	bytes2 "bytes"
	"fmt"

	"github.com/antchfx/xmlquery"
)

// AntchfxXMLFinder represents implementation of XPath from https://github.com/antchfx/xmlquery
type AntchfxXMLFinder struct{}

func NewAntchfxXMLFinder() AntchfxXMLFinder {
	return AntchfxXMLFinder{}
}

func (a AntchfxXMLFinder) Find(expr string, bytes []byte) (any, error) {
	parser, err := xmlquery.Parse(bytes2.NewReader(bytes))
	if err != nil {
		return nil, err
	}

	nodes, err := xmlquery.QueryAll(parser, expr)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 1 {
		return any(nodes[0].InnerText()), nil
	}

	if len(nodes) > 1 {
		results := make([]any, 0, len(nodes))
		for _, node := range nodes {
			results = append(results, node.InnerText())
		}

		return results, nil
	}

	return nil, fmt.Errorf("could not find %s in given XML bytes", expr)
}
