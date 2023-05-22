package wire

import (
	"bytes"
	"io"
	"io/ioutil"
	"scenarios/compressedcoap/ccoapvarint"
)

type HeadersFrameI interface {
	CCoAPFrameI
	Code() uint8
	Token() []byte
	write(w io.Writer) error
}

type _HeadersFrame struct {
	*_CCoAPFrame
	code  uint8
	token []byte
}

func (h *_HeadersFrame) Code() uint8 {
	return h.code
}

func (h *_HeadersFrame) Token() []byte {
	return h.token
}

func (h *_HeadersFrame) write(w io.Writer) error {
	b := ccoapvarint.NewWriter(w)
	h._CCoAPFrame.write(b)

	err := ccoapvarint.WriteBytes(b, uint64(h.code), 1)
	if err != nil {
		return err
	}

	_, err = b.Write(h.token)
	if err != nil {
		return err
	}

	return nil
}

func NewHeaderFrame(code uint8, token []byte) HeadersFrameI {
	length := uint16(5 + len(token))
	return &_HeadersFrame{
		&_CCoAPFrame{length, HEADERS, 0},
		code,
		token,
	}
}

func parseHeadersframe(r *bytes.Reader) (*_HeadersFrame, error) {

	reader := ccoapvarint.NewReader(r)

	code, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	token, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &_HeadersFrame{
		code:  uint8(code),
		token: token,
	}, nil
}

// 	// token, err :=
// 	// frame := &dataFrame{}
// }
