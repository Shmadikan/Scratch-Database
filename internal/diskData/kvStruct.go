package diskdata

import (
	"db/internal/btree_implementation"
	"syscall"
)

type KV struct {
	path string
	fileDescriptor int
	tree btree.Btree
}


func writePages(db *KV) error {
	return nil
}


func updateFile(db *KV) error {
	if err := writePages(db); err != nil {
		return err
	}
	if err := syscall.Fsync(db.fileDescriptor); err != nil {
		return err
	}
	return nil
}




func (db *KV) Open() error
func (db *KV) Get(key []byte) ([]byte, bool) {
	return db.tree.Get(key)
}
func (db *KV) Set(key []byte, val []byte) error {
	db.tree.Insert(key, val)
	return updateFile(db)
}
func (db *KV) Del(key []byte) (bool, error) {
	deleted := db.tree.Delete(key)
	return deleted, updateFile(db)
}