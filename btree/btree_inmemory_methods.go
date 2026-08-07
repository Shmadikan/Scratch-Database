package btree

import (
	"bytes"
	"unsafe"
)

type C struct {
	tree Btree
	ref map[string]string
	pages map[uint64]BNode
}


func newC() *C {
	pages := map[uint64]BNode{}
	return &C{
		tree: Btree{
			get_node: func(ptr uint64) BNode {
				node, _ := pages[ptr]
				return node
			},
			new: func(node BNode) uint64 {
				ptr := uint64(uintptr(unsafe.Pointer(&node[0])))
				pages[ptr] = node
				return ptr
			},
			delete_node: func(ptr uint64) {
				delete(pages, ptr)
			},
		},
		ref: make(map[string]string),
		pages: pages,
	}
}

func (c *C) insert(key string, val string){
	// Дописать что валюха не может больше 3000 быть
	c.tree.Insert([]byte(key), []byte(val))
	c.ref[key] = val
}


func (c *C) delete(key string) bool {

	return c.tree.delete([]byte(key))
	
}


func (c *C) get(key string) string {
	root := c.tree.root_number_page
	node := c.tree.get_node(root)
	index := LookupKeyLE(node, []byte(key))
	
	for ; node.bnode_type() != LEAF; {
		ptr,_ := node.get_kidPointer(index)
		kid_node := c.tree.get_node(ptr)
		node = kid_node
		index = LookupKeyLE(node, []byte(key))
	}
	if (node.bnode_type() == LEAF) {
		if bytes.Compare([]byte(key), node.get_key(index)) == 0 {
			return string(node.get_value(index))
		}
		return ""
	}
	return ""
	
}


