package sqlike

import (
	"errors"
	"sync"
)

var ExpressionNotFound = errors.New("No matching expression was found.")

type likeNode struct {
	val      rune
	wild     bool
	term     bool
	expr     string
	meta     interface{}
	root     bool
	wildnode *likeNode
	children map[rune]*likeNode
}

func (ln likeNode) findExpression(runes []rune) (expr string, meta interface{}, err error) {
	err = ExpressionNotFound
	// Check for edge cases
	if len(runes) == 0 {
		// Only way this happens if runes was empty in the first call
		return "", nil, ExpressionNotFound
	}
	current := runes[0]

	if ln.root {
		if next_node, ok := ln.children[current]; ok {
			expr, meta, err = next_node.findExpression(runes)
			if err == nil {
				return
			}
		}
		if ln.wildnode != nil {
			expr, meta, err = ln.wildnode.findExpression(runes)
			if err == nil {
				return
			}
		}
		return
	}

	// If node is not a root node, we assume that it was called because either
	// it matches the current value or its a wild node

	if len(runes) == 1 {
		if ln.term {
			return ln.expr, ln.meta, nil
		}
		if ln.wild {
			// We are at the end of the text which might be a suffix
			if next_node, ok := ln.children[current]; ok {
				expr, meta, err = next_node.findExpression(runes)
				if err == nil {
					return
				}
			}
		}
		return "", nil, ExpressionNotFound
	}
	next := runes[1]
	if !ln.wild {
		if next_node, ok := ln.children[next]; ok {
			expr, meta, err = next_node.findExpression(runes[1:])
			if err == nil {
				return
			}
		}
		if ln.wildnode != nil {
			expr, meta, err = ln.wildnode.findExpression(runes[1:])
			if err == nil {
				return
			}
		}
		return
	}
	// This is a wild node
	if next_node, ok := ln.children[current]; ok {
		// If wildcard matches nothing
		expr, meta, err = next_node.findExpression(runes)
		if err == nil {
			return
		}
	}
	if next_node, ok := ln.children[next]; ok {
		expr, meta, err = next_node.findExpression(runes[1:])
		if err == nil {
			return
		}
	}
	expr, meta, err = ln.findExpression(runes[1:])
	if err == nil {
		return
	}
	return
}

// Caches like expressions in a trie for pattern matching
type LikeTrie struct {
	mu   *sync.RWMutex
	root *likeNode
}

// Creates a new LikeTrie of Like expressions with an initialized root likeNode.
// TODO: LRU tree limit size is not implemented yet
func NewLikeTrie(size int) *LikeTrie {
	return &LikeTrie{
		root: &likeNode{children: make(map[rune]*likeNode), root: true},
		mu:   &sync.RWMutex{},
	}
}

// Saves a like expression in cache
func (lt *LikeTrie) SaveExpression(like string, meta interface{}) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	runes := []rune(like)
	node := lt.root
	protected := false
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '[' && !protected {
			for j := i + 1; j < len(runes); j++ {
				if runes[j] == ']' {
					protected = true
				}
			}
			if protected {
				continue
			}
		}
		if r == ']' && protected {
			protected = false
			continue
		}
		if r == '%' && !protected {
			if node.wildnode == nil {
				child := &likeNode{
					wild:     true,
					children: make(map[rune]*likeNode),
				}
				node.wildnode = child
			}
			node = node.wildnode
			continue
		}
		if n, ok := node.children[r]; ok {
			node = n
		} else {
			child := &likeNode{
				val:      r,
				children: make(map[rune]*likeNode),
			}
			node.children[r] = child
			node = child
		}
	}
	if node != lt.root {
		node.term = true
		node.expr = like
		node.meta = meta
	}
}

// Looks in the cache to see if there is an expression saved
// which matches the text
// Returns error ExpressionNotFound if no such expression is found
func (lt LikeTrie) FindExpression(text string) (string, interface{}, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	runes := []rune(text)
	return lt.root.findExpression(runes)
}
