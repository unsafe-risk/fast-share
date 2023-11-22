package fastshare

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const headerSize = 8

type Header []byte

var headerPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, headerSize)
		return (*Header)(&b)
	},
}

func NewHeader(name uint32, length uint32) *Header {
	buf := headerPool.Get().(*Header)
	binary.BigEndian.PutUint32((*buf)[:4], name)
	binary.BigEndian.PutUint32((*buf)[4:8], length)
	return buf
}

func (h *Header) Name() uint32 {
	return binary.BigEndian.Uint32((*h)[:4])
}

func (h *Header) Length() uint32 {
	return binary.BigEndian.Uint32((*h)[4:8])
}

func DisposeHeader(b *Header) {
	headerPool.Put(b)
}

func WriteHeaderTo(w io.Writer, h *Header) error {
	_, err := w.Write(*h)
	if err != nil {
		return fmt.Errorf("fastshare.WriteHeaderTo: io.Writer.Write: %w", err)
	}
	return nil
}

func ReadHeaderFrom(r io.Reader, h *Header) error {
	_, err := io.ReadFull(r, *h)
	if err != nil {
		return fmt.Errorf("fastshare.ReadHeaderFrom: io.ReadFull: %w", err)
	}
	return nil
}
