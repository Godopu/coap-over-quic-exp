package converter

import (
	"io"
	"io/ioutil"
	"scenarios/compressedcoap/wire"

	"github.com/plgd-dev/go-coap/v2/message"
)

type msgEncoder interface {
	write(frame wire.CCoAPFrameI) error
	encode(w io.Writer) ([]byte, uint8, error)
}

type frameWriter struct {
	w          io.Writer
	reqWriter  msgEncoder
	respWriter msgEncoder
}

type FrameWriter interface {
	Request(msg *message.Message, upt uint8) ([]byte, uint8, error)
	Response(msg *message.Message, seq uint8) error
}

func NewWriter(w io.Writer) FrameWriter {
	return &frameWriter{
		w:          w,
		reqWriter:  &compressor{},
		respWriter: &responseWriter{},
	}
}

func writeHeader(w msgEncoder, msg *message.Message, seq uint8) error {
	header := wire.NewHeaderFrame(uint8(msg.Code), msg.Token)
	header.SetSEQ(seq)
	err := w.write(header)
	if err != nil {
		return err
	}

	return nil
}

func writeOptions(w msgEncoder, msg *message.Message, seq uint8) error {
	var previousID message.OptionID = 0
	// opts := make([]wire.OptionsFrameI, len(msg.Options))
	for i := 0; i < len(msg.Options); i++ {
		optFrame := wire.NewOptionFrame(
			uint16(msg.Options[i].ID-previousID),
			msg.Options[i].Value,
		)
		optFrame.SetSEQ(seq)
		previousID = msg.Options[i].ID
		err := w.write(optFrame)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeData(w msgEncoder, msg *message.Message, seq uint8) error {
	if msg.Body != nil {
		payload, err := ioutil.ReadAll(msg.Body)
		if err != nil {
			return err
		}

		df := wire.NewDataFrame(payload)
		df.SetSEQ(seq)
		err = w.write(df)
		if err != nil {
			return err
		}
	}

	return nil
}

func (enc *frameWriter) Response(msg *message.Message, seq uint8) error {
	// write header
	err := writeHeader(enc.respWriter, msg, seq)
	if err != nil {
		return err
	}

	// write options
	err = writeOptions(enc.respWriter, msg, seq)
	if err != nil {
		return err
	}

	// write data
	err = writeData(enc.respWriter, msg, seq)
	if err != nil {
		return err
	}

	_, _, err = enc.respWriter.encode(enc.w)
	return err
}

// Request function send request with quic and return compressedToken, seq and error
func (enc *frameWriter) Request(msg *message.Message, upt uint8) ([]byte, uint8, error) {

	// write header
	err := writeHeader(enc.reqWriter, msg, 0)
	if err != nil {
		return nil, 0, err
	}

	// write options
	err = writeOptions(enc.reqWriter, msg, 0)
	if err != nil {
		return nil, 0, err
	}

	err = writeData(enc.reqWriter, msg, 0)
	if err != nil {
		return nil, 0, err
	}

	return enc.reqWriter.encode(enc.w)
}
