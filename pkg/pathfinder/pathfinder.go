// Package pathfinder holds utilities for working with different data formats.
package pathfinder

// PathFinder describes ability to obtain node(s) from data in fixed data format
type PathFinder interface {

	// Find obtains data from bytes according to given expression
	Find(expr string, bytes []byte) (any, error)
}
