package fastshare

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

type FastShareClient struct {
	shmId  int
	length int
	host   string
	buffer []byte

	serverPid int

	conn net.Conn
}

func NewClient(host string) *FastShareClient {
	return &FastShareClient{
		host: host,
	}
}

func (fc *FastShareClient) Connect() error {
	conn, err := net.Dial("tcp", fc.host)
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

	_, err = fc.conn.Read(buf[:])
	if err != nil {
		return fmt.Errorf("fast-share.Connect: net.Conn.Read: %w", err)
	}

	serverPid := binary.BigEndian.Uint64(buf[:])
	fc.serverPid = int(serverPid)

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
	offset := uint32(0)

	buf := [4]byte{}

	for offset < length {
		if _, err := fc.conn.Read(buf[:]); err != nil {
			return 0, 0, fmt.Errorf("fast-share.Receive: net.Conn.Read: %w", err)
		}

		size := binary.BigEndian.Uint32(buf[:])
		if offset+size > length {
			return 0, 0, fmt.Errorf("fast-share.Receive: size(%d) > fc.length(%d)", size, fc.length)
		}

		if _, err := w.Write(fc.buffer[:size]); err != nil {
			return 0, 0, fmt.Errorf("fast-share.Receive: io.Writer.Write: %w", err)
		}

		if err := syscall.Kill(fc.serverPid, syscall.SIGUSR1); err != nil {
			return 0, 0, fmt.Errorf("fast-share.Receive: syscall.Kill: %w", err)
		}

		offset += uint32(size)
	}

	return name, length, nil
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
