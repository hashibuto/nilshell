package ns

import (
	"os"

	"golang.org/x/sys/unix"
)

type FileLock struct {
	filename string
	f        *os.File
}

func NewFileLock(filename string) *FileLock {
	return &FileLock{
		filename: filename,
	}
}

func (fl *FileLock) Lock() error {
	f, err := os.OpenFile(fl.filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	fl.f = f

	err = unix.Flock(int(fl.f.Fd()), unix.LOCK_EX)
	if err != nil {
		return err
	}

	return nil
}

func (fl *FileLock) Unlock() error {
	defer fl.f.Close()

	err := unix.Flock(int(fl.f.Fd()), unix.LOCK_UN)
	if err != nil {
		return err
	}

	return nil
}
