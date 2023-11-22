package fastshare

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"golang.org/x/sys/unix"
)

type FastShareClient struct {
	shmId  int
	length int
	port   int
	buffer []byte

	conn net.Conn
}

func NewClient(port int) *FastShareClient {
	return &FastShareClient{
		port: port,
	}
}

func (fc *FastShareClient) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", fc.port))
	if err != nil {
		return fmt.Errorf("fast-share.Connect: net.Dial: %w", err)
	}
	fc.conn = conn

	buf := [8]byte{}
	_, err = fc.conn.Read(buf[:])
	if err != nil {
		return fmt.Errorf("fast-share.Connect: net.Conn.Read: %w", err)
	}

	id := binary.BigEndian.Uint64(buf[:])
	fc.shmId = int(id)

	return nil
}

func (fc *FastShareClient) Attach() error {
	b, err := unix.SysvShmAttach(fc.shmId, 0, 0)
	if err != nil {
		return fmt.Errorf("fast-share.Attach: unix.SysvShmAttach: %w", err)
	}
	fc.buffer = b
	fc.length = len(b)

	return nil
}

func (fc *FastShareClient) ID() int {
	return fc.shmId
}

func (fc *FastShareClient) Length() int {
	return fc.length
}

func (fc *FastShareClient) Receive(w io.Writer) (uint32, uint32, error) {
	header := NewHeader(0, 0)
	if err := ReadHeaderFrom(fc.conn, header); err != nil {
		return 0, 0, fmt.Errorf("fast-share.Receive: ReadHeaderFrom: %w", err)
	}
	defer DisposeHeader(header)

	name := header.Name()
	length := header.Length()

	// TODO: implement
}

func (fc *FastShareClient) Close() error {
	if err := fc.conn.Close(); err != nil {
		return fmt.Errorf("fast-share.Close: net.Conn.Close: %w", err)
	}

	if err := unix.SysvShmDetach(fc.buffer); err != nil {
		return fmt.Errorf("fast-share.Close: unix.SysvShmDetach: %w", err)
	}

	return nil
}
