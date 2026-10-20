package btree

import (
	"unsafe"
)

type InMemoryStore struct {
	tree   Btree
	values map[string]string
	pages  map[uint64]BNode
	root   BNode
}

func NewInMemoryStore() *InMemoryStore {
	pages := map[uint64]BNode{}
	store := InMemoryStore{
		tree: Btree{
			getNode: func(ptr uint64) BNode {
				node, _ := pages[ptr]
				return node
			},
			newNode: func(node BNode) uint64 {
				ptr := uint64(uintptr(unsafe.Pointer(&node[0])))
				pages[ptr] = node
				return ptr
			},
			deleteNode: func(ptr uint64) {
				delete(pages, ptr)
			},
		},
		values: make(map[string]string),
		pages:  pages,
	}
	store.root = store.tree.getNode(store.tree.rootPageNumber)
	return &store
}

func (s *InMemoryStore) Insert(key string, value string) {
	// Дописать что валюха не может больше 3000 быть
	s.tree.Insert([]byte(key), []byte(value))
	s.values[key] = value
	s.root = s.tree.getNode(s.tree.rootPageNumber)
}

func (s *InMemoryStore) Delete(key string) bool {
	deleted := s.tree.Delete([]byte(key))
	s.root = s.tree.getNode(s.tree.rootPageNumber)
	return deleted
}

func (s *InMemoryStore) Get(key string) string {
	result, found := s.tree.Get([]byte(key))
	if found {
		return string(result)
	}
	return ""
}
