package diskdata

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"golang.org/x/sys/unix"
	b "github.com/Shmadikan/Scratch-Database/internal/btree"	
)

type Store struct {
	Path           string
	fileDescriptor int
	tree           b.Btree
	mmap           struct { // Множество отображение на файл (отображение могут отображаться на разные участки файла)
		total  int // Общий размер всех mmap-ов
		chunks [][]byte
	}
	page struct { // 
		flushed uint64
		temp    [][]byte
	}
}

func writePages(s *Store) error {
	return nil
}


func (s *Store) pageRead(ptr uint64) []byte {
	start := uint64(0)
	for _, chunk := range s.mmap.chunks {
		end := start + uint64(len(chunk) / b.PAGESIZE)
		if ptr <= end {
			offset := b.PAGESIZE * (ptr - start)
			return chunk[offset : offset+b.PAGESIZE]
		}
		start = end
	}
	panic("Doesn't have page with this ptr")
}


func extendMmap(s *Store, size int) error {
	if s.mmap.total > size {
		return nil
	}
	allocation := max(s.mmap.total, 60<<20)
	for s.mmap.total + allocation < size {
		allocation *= 2
	}
	mmap, err := syscall.Mmap(s.fileDescriptor, int64(s.mmap.total), allocation, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("mmap error: %b", mmap)
	}
	s.mmap.chunks = append(s.mmap.chunks, mmap)
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
