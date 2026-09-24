package btree

import (
	"unsafe"
)

type C struct {
	tree Btree
	ref map[string]string
	pages map[uint64]BNode
	root BNode
}


func newC() *C {
	pages := map[uint64]BNode{}
	struct_tree := C{
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
	struct_tree.root = struct_tree.tree.get_node(struct_tree.tree.root_number_page)
	return &struct_tree
}

func (c *C) insert(key string, val string){
	// Дописать что валюха не может больше 3000 быть
	c.tree.Insert([]byte(key), []byte(val))
	c.ref[key] = val
	c.root = c.tree.get_node(c.tree.root_number_page)
}


func (c *C) delete(key string) bool {

	val := c.tree.Delete([]byte(key))
	c.root = c.tree.get_node(c.tree.root_number_page)
	return val
	
}


func (c *C) get(key string) string {
	result, flag := c.tree.Get([]byte(key))
	if flag {
		return string(result)
	}
	return ""
}


