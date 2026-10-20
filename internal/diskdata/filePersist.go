package diskdata

import (
	"fmt"
	"os"
	"path"
	"syscall"
)

func createFileSync(filename string) (int, error) {
	// Создание файла директории с мгновенным сбросом буфера
	flags := os.O_RDONLY | syscall.O_DIRECTORY
	dirFD, err := syscall.Open(path.Dir(filename), flags, 0o644)
	if err != nil {
		return -1, fmt.Errorf("%w", err)
	}
	flags = os.O_CREATE | os.O_RDWR
	fd, err := syscall.Openat(dirFD, filename, flags, 0o644)
	if err != nil {
		return -1, fmt.Errorf("%w", err)
	}
	if err := syscall.Fsync(dirFD); err != nil {
		_ = syscall.Close(fd)
		return -1, fmt.Errorf("%w", err)
	}
	return fd, nil
}
