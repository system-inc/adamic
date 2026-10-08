//go:build linux || darwin

package flattree

import (
	"os"
	"syscall"
)

// Mapping owns the lifetime of a read-only Unix MAP_PRIVATE view. Reader must
// never be used after Close, and Close must not race any reader.
type Mapping struct {
	Reader Reader
	bytes  []byte
}

func Map(path string) (*Mapping, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() == 0 {
		return nil, syscall.EINVAL
	}
	b, err := syscall.Mmap(int(f.Fd()), 0, int(info.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, err
	}
	r, err := Open(b)
	if err != nil {
		_ = syscall.Munmap(b)
		return nil, err
	}
	return &Mapping{Reader: r, bytes: b}, nil
}
func (m *Mapping) Close() error {
	err := syscall.Munmap(m.bytes)
	m.bytes = nil
	m.Reader = Reader{}
	return err
}
