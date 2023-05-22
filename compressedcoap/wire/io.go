package wire

import (
	"errors"
	"io"
)

func Write(w io.Writer, frame CCoAPFrameI) error {
	hf, ok := frame.(HeadersFrameI)
	if ok {
		hf.write(w)
		return nil
	}

	of, ok := frame.(OptionsFrameI)
	if ok {
		of.write(w)
		return nil
	}

	df, ok := frame.(DataFrameI)
	if ok {
		df.write(w)
		return nil
	}

	return errors.New("wrong type error")

}
