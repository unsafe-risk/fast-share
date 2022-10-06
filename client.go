package fastshare

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type FastShareClient struct {
	id     int
	length int
	buffer []byte
}

func Connect(id int) (*FastShareClient, error) {
	fs := &FastShareClient{}
	b, err := unix.SysvShmAttach(id, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("fast-share.Connect: unix.SysvShmAttach: %w", err)
	}
	fs.id = id
	fs.buffer = b
	fs.length = len(b)
	return fs, nil
}

func (fs *FastShareClient) ID() int {
	return fs.id
}

func (fs *FastShareClient) Length() int {
	return fs.length
}
