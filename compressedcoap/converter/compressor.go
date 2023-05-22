package converter

import (
	"errors"
	"io"
	"log"
	"scenarios/compressedcoap/wire"
)

// import "scenarios/compressedcoap/wire"

type compressor struct {
	latestHEADERS wire.HeadersFrameI
	latestOPTIONS []wire.OptionsFrameI
	SEQ           uint8
	frames        []wire.CCoAPFrameI
}

func (c *compressor) write(frame wire.CCoAPFrameI) error {
	// routine for check error is not implemented now.

	c.frames = append(c.frames, frame)
	return nil
}

func (c *compressor) checkOptionCompressibility(optframes []wire.CCoAPFrameI) bool {

	// for latest
	if len(c.latestOPTIONS) != len(optframes) {
		return false
	}

	for i, frame := range optframes {
		optframe, ok := frame.(wire.OptionsFrameI)
		if !ok {
			log.Println("detect not options frame during check compressibility of options")
			return false
		}

		if optframe.OptionDelta() != c.latestOPTIONS[i].OptionDelta() ||
			string(optframe.OptionValue()) != string(c.latestOPTIONS[i].OptionValue()) {
			return false
		}
	}

	return true
}

func (c *compressor) encode(w io.Writer) ([]byte, uint8, error) {
	// todo
	// send empty frame if we don't have any frame to send
	defer func() {
		c.frames = nil
	}()
	var err error
	// length := len(c.frames)

	// set fin flag to last frame
	c.frames[len(c.frames)-1].SetFin()

	// get start index of data frame
	var dataFrameIdx = len(c.frames)
	dataframe, ok := c.frames[len(c.frames)-1].(wire.DataFrameI)
	if ok {
		dataFrameIdx--
	}

	isSameOption := c.checkOptionCompressibility(c.frames[1:dataFrameIdx])
	// encode header frame
	hf, ok := c.frames[0].(wire.HeadersFrameI)
	if !ok {
		return nil, 0, errors.New("first frame should be HEADERS")
	}

	if c.latestHEADERS == nil || c.latestHEADERS.Code() != hf.Code() || !isSameOption || c.SEQ == 255 {
		c.SEQ = 0
		c.latestHEADERS = hf
		err = wire.Write(w, c.frames[0])
		if err != nil {
			return nil, 0, err
		}
	} else {
		c.SEQ++
		// hf.SetSEQ(c.SEQ)
	}

	// Add token in header to token table
	// if hf.Code()&0xE0 == 0 {
	// 	AddReqToken(c.latestHEADERS.Token(), hf.Token(), c.SEQ)
	// }

	// encode option
	if (c.latestOPTIONS == nil || len(c.latestOPTIONS) == 0) && dataFrameIdx > 1 || !isSameOption || c.SEQ == 0 {
		// it is not possible to compress options frames
		c.latestOPTIONS = c.latestOPTIONS[:0] // flush options

		// copy
		for _, optFrame := range c.frames[1:dataFrameIdx] {
			optFrame.SetSEQ(c.SEQ)
			c.latestOPTIONS = append(c.latestOPTIONS, optFrame.(wire.OptionsFrameI))
			err = wire.Write(w, optFrame)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	if dataframe != nil {
		dataframe.SetSEQ(c.SEQ)
		wire.Write(w, dataframe)
		if err != nil {
			return nil, 0, err
		}
	}

	return c.latestHEADERS.Token(), c.SEQ, nil
}

// // func (c *compressor)
