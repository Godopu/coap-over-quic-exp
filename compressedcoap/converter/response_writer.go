package converter

import (
	"errors"
	"io"
	"scenarios/compressedcoap/wire"
)

type responseWriter struct {
	frames []wire.CCoAPFrameI
}

func (rw *responseWriter) write(frame wire.CCoAPFrameI) error {
	rw.frames = append(rw.frames, frame)
	return nil
}

func (rw *responseWriter) encode(w io.Writer) ([]byte, uint8, error) {
	defer func() {
		rw.frames = nil
	}()

	hf, ok := rw.frames[0].(wire.HeadersFrameI)
	if !ok {
		return nil, 0, errors.New("first frame should be HEADERS")
	}

	rw.frames[len(rw.frames)-1].SetFin()

	for _, frame := range rw.frames {
		err := wire.Write(w, frame)
		if err != nil {
			return nil, 0, err
		}
	}

	return hf.Token(), hf.Seq(), nil
}
