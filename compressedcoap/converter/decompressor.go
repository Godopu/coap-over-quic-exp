package converter

import (
	"errors"
	"scenarios/compressedcoap/wire"
)

type decompressor struct {
	latestHEADERS wire.HeadersFrameI
	latestOPTIONS []wire.OptionsFrameI
	SEQ           uint8
}

func (d *decompressor) decompressHeadersFrame(seq uint8) (wire.HeadersFrameI, error) {
	if d.latestHEADERS == nil {
		return nil, errors.New("latestHeader is nil")
	}

	header := wire.NewHeaderFrame(d.latestHEADERS.Code(), d.latestHEADERS.Token())
	header.SetSEQ(seq)

	return header, nil
}

func (d *decompressor) decompressOptionsFrame(seq uint8) ([]wire.OptionsFrameI, error) {
	for _, opt := range d.latestOPTIONS {
		opt.SetSEQ(seq)
	}
	return d.latestOPTIONS, nil
}
