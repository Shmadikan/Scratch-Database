package diskdata

import (
	"github.com/Shmadikan/Scratch-Database/internal/btree"
	"syscall"
)

type Store struct {
	Path           string
	fileDescriptor int
	tree           btree.Btree
}

func writePages(s *Store) error {
	return nil
}

func syncFile(s *Store) error {
	if err := writePages(s); err != nil {
		return err
	}
	if err := syscall.Fsync(s.fileDescriptor); err != nil {
		return err
	}
	return nil
}

func (s *Store) Open() error {
	fd, err := createFileSync(s.Path)
	if err != nil {
		return err
	}
	s.fileDescriptor = fd
	return nil
}

func (s *Store) Get(key []byte) ([]byte, bool) {
	return s.tree.Get(key)
}

func (s *Store) Set(key []byte, value []byte) error {
	s.tree.Insert(key, value)
	return syncFile(s)
}

func (s *Store) Delete(key []byte) (bool, error) {
	deleted := s.tree.Delete(key)
	return deleted, syncFile(s)
}
