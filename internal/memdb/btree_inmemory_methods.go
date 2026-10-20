package btree

import (
	"unsafe"

	b "github.com/Shmadikan/Scratch-Database/internal/btree"
)

type inMemoryStore struct {
	tree   b.Btree
	values map[string]string
	pages  map[uint64]b.BNode
	root   b.BNode
}

func newInMemoryStore() *inMemoryStore {
	pages := map[uint64]b.BNode{}
	store := inMemoryStore{
		tree: b.Btree{
			GetNode: func(ptr uint64) b.BNode {
				node, _ := pages[ptr]
				return node
			},
			NewNode: func(node b.BNode) uint64 {
				ptr := uint64(uintptr(unsafe.Pointer(&node[0])))
				pages[ptr] = node
				return ptr
			},
			DeleteNode: func(ptr uint64) {
				delete(pages, ptr)
			},
		},
		values: make(map[string]string),
		pages:  pages,
	}
	store.root = store.tree.GetNode(store.tree.RootPageNumber)
	return &store
}

func (s *inMemoryStore) Insert(key string, value string) {
	// Дописать что валюха не может больше 3000 быть
	s.tree.Insert([]byte(key), []byte(value))
	s.values[key] = value
	s.root = s.tree.GetNode(s.tree.RootPageNumber)
}

func (s *inMemoryStore) Delete(key string) bool {
	deleted := s.tree.Delete([]byte(key))
	s.root = s.tree.GetNode(s.tree.RootPageNumber)
	return deleted
}

func (s *inMemoryStore) Get(key string) string {
	result, found := s.tree.Get([]byte(key))
	if found {
		return string(result)
	}
	return ""
}
