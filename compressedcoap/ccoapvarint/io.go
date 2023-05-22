package ccoapvarint

import (
	"bytes"
	"io"
)

// io.Reader can't read byte, so we should override ReadByte method to convert io.Reader to io.byteReader
type Reader interface {
	io.ByteReader
	io.Reader
	ReadBytesAsInt64(n int) (uint64, error)
	ReadBytes(n uint64) (*bytes.Reader, error)
}

type byteReader struct {
	io.Reader
}

func NewReader(r io.Reader) Reader {
	if r, ok := r.(Reader); ok {
		return r
	}
	return &byteReader{r}
}

func (r *byteReader) ReadByte() (byte, error) {
	var b [1]byte
	_, err := r.Reader.Read(b[:])
	return b[0], err
}

func (r *byteReader) ReadBytesAsInt64(n int) (uint64, error) {
	var b [1]byte
	var ret uint64 = 0

	for i := 0; i < n; i++ {
		_, err := r.Reader.Read(b[:])
		if err != nil {
			return 0, nil
		}
		ret += uint64(b[0]) << (8 * ((n - 1) - i))
	}

	return ret, nil
}

func (r *byteReader) ReadBytes(n uint64) (*bytes.Reader, error) {
	buffer := make([]byte, n)
	total := uint64(0)

	for total < n {
		s, err := r.Reader.Read(buffer)
		if err != nil {
			return nil, err
		}
		total += uint64(s)
	}

	return bytes.NewReader(buffer), nil
}

type Writer interface {
	io.ByteWriter
	io.Writer
}

type byteWriter struct {
	io.Writer
}

func NewWriter(w io.Writer) Writer {
	if w, ok := w.(Writer); ok {
		return w
	}
	return &byteWriter{w}
}

func (w *byteWriter) WriteByte(c byte) error {
	_, err := w.Writer.Write([]byte{c})
	return err
}

func WriteBytes(w Writer, v uint64, len int) error {
	b := make([]byte, len)

	for i := 0; i < len; i++ {
		b[len-i-1] = uint8(v >> (8 * i))
	}

	_, err := w.Write(b)
	return err
	// w.Write([]byte{uint8(i >> 24), uint8(i >> 16), uint8(i >> 8), uint8(i)})
	// if i <= maxVarInt1 {
	// 	w.WriteByte(uint8(i))
	// } else if i <= maxVarInt2 {
	// 	w.Write([]byte{uint8(i>>8) | 0x40, uint8(i)})
	// } else if i <= maxVarInt4 {
	// 	w.Write([]byte{uint8(i>>24) | 0x80, uint8(i >> 16), uint8(i >> 8), uint8(i)})
	// } else if i <= maxVarInt8 {
	// 	w.Write([]byte{
	// 		uint8(i>>56) | 0xc0, uint8(i >> 48), uint8(i >> 40), uint8(i >> 32),
	// 		uint8(i >> 24), uint8(i >> 16), uint8(i >> 8), uint8(i),
	// 	})
	// } else {
	// 	panic(fmt.Sprintf("%#x doesn't fit into 62 bits", i))
	// }
}
