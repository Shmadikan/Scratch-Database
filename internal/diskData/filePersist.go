package diskdata

import (
	"fmt"
	"os"
	"path"
	"syscall"
)


func createFileSync(file string) (int, error) {
	// Создание файла директории с мгновенным сбросом буфера
	flags := os.O_RDONLY | syscall.O_DIRECTORY
	dirfd, err := syscall.Open(path.Dir(file), flags, 0o644)
	if err != nil {
		return -1, fmt.Errorf(err.Error())
	}
	flags = os.O_CREATE | os.O_RDWR
	fd, err := syscall.Openat(dirfd, file, flags, 0o644)
	if err != nil {
		return -1, fmt.Errorf(err.Error())
	}
	if err := syscall.Fsync(dirfd); err != nil {
		_ = syscall.Close(fd)
		return -1, fmt.Errorf(err.Error())
	}
	return fd, nil
}