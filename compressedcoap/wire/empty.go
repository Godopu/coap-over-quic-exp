package wire

import "io"

type EmptyFrameI interface {
	CCoAPFrameI
	write(b io.Writer) error
}

type _OptionFrame struct {
	*_CCoAPFrame
	extendedLength uint64
	payload        []byte
}
