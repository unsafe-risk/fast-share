package fastshare

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"

	"golang.org/x/sys/unix"
)

type FastShareServer struct {
	id     int
	length int
	buffer []byte

	pid int

	client    net.Conn
	connected atomic.Bool
}

func NewServer(key int, bitSize int) (*FastShareServer, error) {
	if bitSize < 32 {
		return nil, fmt.Errorf("fast-share.NewServer: length must be at least 8 bytes")
	}

	fs := &FastShareServer{}

	id, err := unix.SysvShmGet(key, bitSize, unix.IPC_CREAT|0o660)
	if err != nil {
		return nil, fmt.Errorf("fast-share.NewServer: unix.SysvShmGet: %w", err)
	}
	fs.id = id

	b, err := unix.SysvShmAttach(id, 0, 0)
	if err != nil {
		unix.SysvShmCtl(id, unix.IPC_RMID, nil)
		return nil, fmt.Errorf("fast-share.NewServer: unix.SysvShmAttach: %w", err)
	}
	fs.buffer = b
	fs.length = len(b)

	for i := range fs.buffer {
		fs.buffer[i] = 0
	}

	fs.pid = unix.Getpid()

	return fs, nil
}

func (fs *FastShareServer) Listen(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return fmt.Errorf("fast-share.Listen: net.Listen: %w", err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			return fmt.Errorf("fast-share.Listen: net.Listener.Accept: %w", err)
		}

		go func() {
			if fs.connected.Swap(true) {
				conn.Close()
				return
			}

			buf := [8]byte{}
			binary.BigEndian.PutUint64(buf[:], uint64(fs.id))
			_, err := conn.Write(buf[:])
			if err != nil {
				return
			}

			fs.client = conn
		}()
	}
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

func (fs *FastShareServer) Send(name uint32, data []byte) error {
	if !fs.connected.Load() || fs.client == nil {
		return fmt.Errorf("fast-share.Send: not connected")
	}

	h := NewHeader(name, uint32(len(data)))
	if err := WriteHeaderTo(fs.client, h); err != nil {
		return fmt.Errorf("fast-share.Send: WriteHeaderTo: %w", err)
	}
	defer DisposeHeader(h)

	offset := 0
	currentBufSize := [8]byte{}

	for offset < len(data) {
		last := offset + fs.length
		if last > len(data) {
			last = len(data)
		}

		copy(fs.buffer, data[offset:last])

		binary.BigEndian.PutUint64(currentBufSize[:], uint64(last-offset))

		if _, err := fs.client.Write(currentBufSize[:]); err != nil {
			return fmt.Errorf("fast-share.Send: net.Conn.Write: %w", err)
		}

		if _, err := fs.client.Read(currentBufSize[:]); err != nil {
			return fmt.Errorf("fast-share.Send: net.Conn.Read: %w", err)
		}

		offset = last
	}

	return nil
}

func (fs *FastShareServer) Disconnect() error {
	defer fs.connected.Store(false)

	if err := fs.client.Close(); err != nil {
		return fmt.Errorf("fast-share.Disconnect: net.Conn.Close: %w", err)
	}

	return nil
}
