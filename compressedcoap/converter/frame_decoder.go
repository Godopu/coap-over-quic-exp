package converter

import (
	"bytes"
	"errors"
	"io"
	"scenarios/compressedcoap/wire"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
)

type decoder struct {
	r     io.Reader
	dcmps *decompressor
}

type Decoder interface {
	Decode(msg *message.Message, upt *uint8) (uint8, error)
}

func NewDecoder(r io.Reader) Decoder {
	return &decoder{r, &decompressor{}}
}

func (dec *decoder) decodeFromSecondFrame(msg *message.Message) error {
	var previousOptID message.OptionID = 0

	if len(msg.Options) != 0 {
		previousOptID = msg.Options[0].ID
	}

	for {
		frame, err := wire.ParseFrame(dec.r)
		if err != nil {
			return err
		}

		switch frame.Type() {
		case wire.HEADERS:
			return errors.New("duplicated headers error")

		case wire.OPTIONS:
			option, ok := frame.(wire.OptionsFrameI)
			if !ok {
				return errors.New("error is occured on type converting")
			}
			opt := message.Option{
				ID:    previousOptID + message.OptionID(option.OptionDelta()),
				Value: option.OptionValue(),
			}
			msg.Options = append(msg.Options, opt)
			previousOptID += message.OptionID(option.OptionDelta())

			if option.Seq() == 0 {
				dec.dcmps.latestOPTIONS = append(dec.dcmps.latestOPTIONS, option)
			}

		case wire.DATA:
			data, ok := frame.(wire.DataFrameI)
			if !ok {
				return errors.New("error is occured on type converting")
			}
			msg.Body = bytes.NewReader(data.Payload())
			return nil
		}

		if frame.IsFin() {
			break
		}
	}

	return nil
}

func (dec *decoder) Decode(msg *message.Message, upt *uint8) (uint8, error) {
	// decode first frame
	var seq uint8 = 0
	frame, err := wire.ParseFrame(dec.r)
	if err != nil {
		return 0, err
	}
	*upt = frame.GetUnderlyingProtocol()

	var header wire.HeadersFrameI
	ok := false

	switch frame.Type() {
	case wire.HEADERS:
		header, ok = frame.(wire.HeadersFrameI)
		if !ok {
			return 0, errors.New("error is occured on type converting")
		}

		seq = header.Seq()
		msg.Code = codes.Code(header.Code())
		msg.Token = header.Token()

		dec.dcmps.latestHEADERS = header

		err := dec.decodeFromSecondFrame(msg)
		if err != nil {
			return 0, err
		}

	case wire.OPTIONS:
		// decompress headers frame
		seq = frame.Seq()
		header, err = dec.dcmps.decompressHeadersFrame(frame.Seq())
		if err != nil {
			return 0, err
		}

		msg.Code = codes.Code(header.Code())
		msg.Token = header.Token()

		// parse first option
		option, ok := frame.(wire.OptionsFrameI)
		if !ok {
			return 0, errors.New("error is occured on type converting")
		}
		opt := message.Option{
			ID:    message.OptionID(int(option.OptionDelta())),
			Value: option.OptionValue(),
		}
		msg.Options = append(msg.Options, opt)

		err := dec.decodeFromSecondFrame(msg)
		if err != nil {
			return 0, err
		}
	case wire.DATA:
		// decompress header frame
		seq = frame.Seq()
		header, err = dec.dcmps.decompressHeadersFrame(frame.Seq())
		if err != nil {
			return 0, err
		}

		msg.Code = codes.Code(header.Code())
		msg.Token = header.Token()

		// decompress options frame
		options, err := dec.dcmps.decompressOptionsFrame(frame.Seq())
		if err != nil {
			return 0, err
		}

		opts := make(message.Options, len(options))
		previousOptID := 0
		for i, option := range options {
			opts[i] = message.Option{
				ID:    message.OptionID(previousOptID) + message.OptionID(option.OptionDelta()),
				Value: option.OptionValue(),
			}

			previousOptID += int(option.OptionDelta())
		}

		msg.Options = opts

		// parse frame
		data, ok := frame.(wire.DataFrameI)
		if !ok {
			return 0, errors.New("error is occured on type converting")
		}
		msg.Body = bytes.NewReader(data.Payload())
	}

	return seq, nil
}
