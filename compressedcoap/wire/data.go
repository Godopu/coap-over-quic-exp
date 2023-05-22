package wire

import (
	"bytes"
	"io"
	"io/ioutil"
	"scenarios/compressedcoap/ccoapvarint"
)

// import "bytes"

type DataFrameI interface {
	CCoAPFrameI
	ExtendedLength() uint64
	Payload() []byte
	write(b io.Writer) error
}

type _DataFrame struct {
	*_CCoAPFrame
	extendedLength uint64
	payload        []byte
}

func (d *_DataFrame) ExtendedLength() uint64 {
	return uint64(d.extendedLength)
}

func (d *_DataFrame) Payload() []byte {
	return d.payload
}

func (d *_DataFrame) write(w io.Writer) error {
	b := ccoapvarint.NewWriter(w)
	d._CCoAPFrame.write(b)

	if d.extendedLength != 0 {
		err := ccoapvarint.WriteBytes(b, uint64(d.extendedLength), 8)
		if err != nil {
			return err
		}
	}
	_, err := b.Write(d.payload)
	if err != nil {
		return err
	}

	return nil
}

func NewDataFrame(payload []byte) DataFrameI {
	var length uint16 = 0
	var extLength uint64 = 0

	if len(payload) > 0xffff-4 {
		length = 0
		extLength = uint64(4 + 8 + len(payload))
	} else {
		length = uint16(4 + len(payload))
		extLength = 0
	}

	return &_DataFrame{
		&_CCoAPFrame{length, DATA, 0},
		uint64(extLength),
		payload,
	}
}

func parseDataframe(r io.Reader, length uint16) (*_DataFrame, error) {

	reader := ccoapvarint.NewReader(r)

	var extendedLength uint64 = 0
	var err error
	if length == 0 {
		extendedLength, err = reader.ReadBytesAsInt64(8)
		if err != nil {
			return nil, err
		}
	}

	var payloadReader *bytes.Reader = nil
	if length == 0 {
		payloadReader, err = reader.ReadBytes(extendedLength - 12)
	} else {
		payloadReader, err = reader.ReadBytes(uint64(length - 4))
	}
	if err != nil {
		return nil, err
	}

	payload, err := ioutil.ReadAll(payloadReader)
	if err != nil {
		return nil, err
	}

	return &_DataFrame{
		extendedLength: extendedLength,
		payload:        payload,
	}, nil
}

// func parseDataframe(r *bytes.Reader) (DataFrameI, error) {
// 	codeByte, err := r.ReadByte()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// token, err :=
// 	// frame := &dataFrame{}
// }

// func Write(b *bytes.Buffer) error {

// }
