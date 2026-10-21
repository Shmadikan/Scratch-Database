package diskdata

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"golang.org/x/sys/unix"
	b "github.com/Shmadikan/Scratch-Database/internal/btree"	
)

const DB_SIG = "Exactly16BytesIn"


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
	failed bool
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
	s.mmap.total += len(mmap)
	return nil
}


func (s *Store) pageAppend(node []byte) uint64 {
	s.page.temp = append(s.page.temp, node)
	return uint64(s.page.flushed + uint64(len(s.page.temp)))
}


func writePages(s *Store) error {
	sum := 0
	for _, temp := range s.page.temp {
		sum += len(temp)
	}
	err := extendMmap(s, s.mmap.total + sum)
	if err != nil {
		return fmt.Errorf("erro while extending pages: %s", err)
	}
	_, err = unix.Pwritev(s.fileDescriptor, s.page.temp, int64(s.page.flushed * b.PAGESIZE))
	if err != nil {
		return fmt.Errorf("error while writing in file: %s", err)
	}
	s.page.flushed += uint64(len(s.page.temp))
	s.page.temp = s.page.temp[:0]
	return nil
}


func saveMeta(s *Store) []byte {
	data := make([]byte, 32)
	copy(data, []byte(DB_SIG))
	binary.LittleEndian.PutUint64(data[16:], s.tree.RootPageNumber)
	binary.LittleEndian.PutUint64(data[24:], s.page.flushed)
	return data
}


func loadMeta(s *Store, data []byte) {
	s.tree.RootPageNumber = binary.LittleEndian.Uint64(data[16:])
	s.page.flushed = binary.LittleEndian.Uint64(data[24:])
}


func readRoot(s *Store, fileSize int64) error {
	if fileSize == 0 {
		s.page.flushed = 1
		return nil
	}
	startPage := s.pageRead(0)
	loadMeta(s, startPage)
	if binary.LittleEndian.Uint64(startPage[16:]) != s.tree.RootPageNumber {
		return fmt.Errorf("Corrupted root page. Rootptr is Corrupted")
	}
	if binary.LittleEndian.Uint64(startPage[24:]) != s.page.flushed {
		return fmt.Errorf("Corrupted root page. Rootptr is Corrupted")
	}
	return nil
}


func updateRoot(s *Store) error {
	_, err := syscall.Pwrite(s.fileDescriptor, saveMeta(s), 0)
	if err != nil {
		return fmt.Errorf("Error while writing: %s", err)
	}
	return nil
}


func updateFile(s *Store) error {
	err := writePages(s)
	if err != nil {
		return fmt.Errorf("Error with update file: %s", err)
	}
	if err = syscall.Fsync(s.fileDescriptor); err != nil {
		return fmt.Errorf("Error with fsync:%s", err)
	}
	if err = updateRoot(s); err != nil {
		return fmt.Errorf("Error with update file:%s", err)
	}
	if err = syscall.Fsync(s.fileDescriptor); err != nil {
		return fmt.Errorf("Error with fsync:%s", err)
	}
	return nil
}


func updateOrRevert(s *Store, meta[]byte) error {
	if s.failed {
		if err := updateRoot(s); err != nil {
		return err
	}
		if err := syscall.Fsync(s.fileDescriptor); err != nil{
		return err
		}
		loadMeta(s, meta)
		s.failed = false
	}
	err := updateFile(s)
	if err != nil {
		loadMeta(s, meta)
		s.page.temp = s.page.temp[:0]
		s.failed = true
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
	return updateFile(s)
}

func (s *Store) Delete(key []byte) (bool, error) {
	deleted := s.tree.Delete(key)
	return deleted, updateFile(s)
}
