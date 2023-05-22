package wire

import (
	"bytes"
	"io"
	"io/ioutil"
	"scenarios/compressedcoap/ccoapvarint"
)

type OptionsFrameI interface {
	CCoAPFrameI
	OptionDelta() uint16
	OptionValue() []byte
	write(w io.Writer) error
}

type _OptionsFrame struct {
	*_CCoAPFrame
	optionDelta uint16
	optionValue []byte
}

func (opt *_OptionsFrame) OptionDelta() uint16 {
	return opt.optionDelta
}

func (opt *_OptionsFrame) OptionValue() []byte {
	return opt.optionValue
}

func (opt *_OptionsFrame) write(w io.Writer) error {
	b := ccoapvarint.NewWriter(w)
	opt._CCoAPFrame.write(b)

	err := ccoapvarint.WriteBytes(b, uint64(opt.optionDelta), 2)
	if err != nil {
		return err
	}

	_, err = b.Write(opt.optionValue)
	if err != nil {
		return err
	}

	return nil
}

func NewOptionFrame(optionDelta uint16, optionValue []byte) OptionsFrameI {
	length := uint16(6 + len(optionValue))
	return &_OptionsFrame{
		&_CCoAPFrame{length, OPTIONS, 0},
		optionDelta,
		optionValue,
	}
}

func parseOptionsFrame(r *bytes.Reader) (*_OptionsFrame, error) {

	reader := ccoapvarint.NewReader(r)

	optionDelta, err := reader.ReadBytesAsInt64(2)
	if err != nil {
		return nil, err
	}

	optionValue, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &_OptionsFrame{
		optionDelta: uint16(optionDelta),
		optionValue: optionValue}, nil

}
