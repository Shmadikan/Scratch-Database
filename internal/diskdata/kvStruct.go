package diskdata

import (
	"syscall"

	"github.com/Shmadikan/Scratch-Database/internal/btree"
)

type DB struct {
	path           string
	fileDescriptor int
	tree           btree.Btree
}

func writePages(db *DB) error {
	return nil
}

func syncFile(db *DB) error {
	if err := writePages(db); err != nil {
		return err
	}
	if err := syscall.Fsync(db.fileDescriptor); err != nil {
		return err
	}
	return nil
}

func (db *DB) Open() error {
	fd, err := createFileSync(db.path)
	if err != nil {
		return err
	}
	db.fileDescriptor = fd
	return nil
}

func (db *DB) Get(key []byte) ([]byte, bool) {
	return db.tree.Get(key)
}

func (db *DB) Set(key []byte, value []byte) error {
	db.tree.Insert(key, value)
	return syncFile(db)
}

func (db *DB) Delete(key []byte) (bool, error) {
	deleted := db.tree.Delete(key)
	return deleted, syncFile(db)
}
