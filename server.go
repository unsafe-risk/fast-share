package fastshare

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

type FastShareServer struct {
	id     int
	length int
	buffer []byte

	pid  int
	conn net.Conn
}

func Listen(key int, length int, socketFileName string) (*FastShareServer, error) {
	if length < 32 {
		return nil, fmt.Errorf("fast-share.Listen: length must be at least 8 bytes")
	}

	fs := &FastShareServer{}

	id, err := unix.SysvShmGet(key, length, unix.IPC_CREAT|0o660)
	if err != nil {
		return nil, fmt.Errorf("fast-share.Listen: unix.SysvShmGet: %w", err)
	}
	fs.id = id

	b, err := unix.SysvShmAttach(id, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("fast-share.New: unix.SysvShmAttach: %w", err)
	}
	fs.buffer = b
	fs.length = len(b)

	for i := range fs.buffer {
		fs.buffer[i] = 0
	}

	fs.pid = unix.Getpid()

	return fs, nil
}

func (fs *FastShareServer) ID() int {
	return fs.id
}

func (fs *FastShareServer) Length() int {
	return fs.length
}

func (fs *FastShareServer) Close() error {
	if err := unix.SysvShmDetach(fs.buffer); err != nil {
		return fmt.Errorf("fast-share.Close: unix.SysvShmDetach: %w", err)
	}

	_, err := unix.SysvShmCtl(fs.id, unix.IPC_RMID, nil)
	if err != nil {
		return fmt.Errorf("fast-share.Close: unix.SysvShmCtl: %w", err)
	}

	return nil
}

func (fs *FastShareServer) PID() int {
	return fs.pid
}
