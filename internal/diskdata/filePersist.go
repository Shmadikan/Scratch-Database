package diskdata

import (
	"os"
	"path"
	"syscall"
)

func createFileSync(filePath string) (int, error) {
	// Создание файла директории с мгновенным сбросом буфера
	flags := os.O_RDONLY | syscall.O_DIRECTORY
	dirFd, err := syscall.Open(path.Dir(filePath), flags, 0o644)
	if err != nil {
		return -1, err
	}
	flags = os.O_CREATE | os.O_RDWR
	fd, err := syscall.Openat(dirFd, filePath, flags, 0o644)
	if err != nil {
		return -1, err
	}
	if err := syscall.Fsync(dirFd); err != nil {
		_ = syscall.Close(fd)
		return -1, err
	}
	return fd, nil
}
