package compressedcoap

import (
	"errors"
	"sync"
)

type compressedTknPair struct {
	CompressedToken []byte
	Seq             uint8
}

type tokenMapper struct {

	// newToken, SEQ -> originToken
	latestTkn      []byte
	encodedMsgTkns map[uint64][][]byte
	// newToken -> originToken, SEQ
	decodedMsgTkns map[uint64]*compressedTknPair
	encodedMutex   sync.Mutex
	decodedMutex   sync.Mutex
}

func (mapper *tokenMapper) pushEncodedMsgTkn(originToken, compressedToken []byte, seq uint8) error {

	// add token when message is issued
	// if wrong order seq is received, return error
	if mapper.encodedMsgTkns[getHashValue(compressedToken)] == nil && seq != 0 {
		return errors.New("wrong order seq is received")
	}

	mapper.encodedMutex.Lock()
	defer mapper.encodedMutex.Unlock()
	if seq == 0 {
		mapper.latestTkn = compressedToken
		mapper.encodedMsgTkns[getHashValue(compressedToken)] = make([][]byte, 0, 256)
	}

	// if new token is received, add the token to tknTable
	mapper.encodedMsgTkns[getHashValue(compressedToken)] = append(mapper.encodedMsgTkns[getHashValue(compressedToken)], originToken)

	return nil
}

func (mapper *tokenMapper) popEncodedMsgTkn(cmpTkn []byte, seq uint8) ([]byte, error) {
	defer func() {
		if seq == 255 ||
			getHashValue(cmpTkn) != getHashValue(mapper.latestTkn) && int(seq) >= len(mapper.encodedMsgTkns[getHashValue(cmpTkn)]) {
			delete(mapper.encodedMsgTkns, getHashValue(cmpTkn))
		}
	}()

	originTkns, ok := mapper.encodedMsgTkns[getHashValue(cmpTkn)]
	if !ok {
		return nil, errors.New("wrong comressed token is sended")
	}

	if len(originTkns) <= int(seq) {
		return nil, errors.New("wrong sequence number is sended")
	}

	// mapper.encodedMutex.Lock()
	// defer mapper.encodedMutex.Unlock()
	// delete(mapper.encodedMsgTkns, getHashValue(cmpTkn))

	return originTkns[seq], nil
}

func (mapper *tokenMapper) pushDecodedMsgTkn(newTkn, compressTkn []byte, seq uint8) error {
	_, ok := mapper.encodedMsgTkns[getHashValue(newTkn)]
	if ok {
		return errors.New("duplicate token detection")
	}

	mapper.decodedMutex.Lock()
	defer mapper.decodedMutex.Unlock()
	mapper.decodedMsgTkns[getHashValue(newTkn)] = &compressedTknPair{
		CompressedToken: compressTkn,
		Seq:             seq,
	}

	return nil
}

func (mapper *tokenMapper) popDecodedMsgTkn(newTkn []byte) ([]byte, uint8, error) {
	cmpTkn, ok := mapper.decodedMsgTkns[getHashValue(newTkn)]
	if !ok {
		return nil, 0, errors.New("cannot find token pair corresponded to received token")
	}

	mapper.decodedMutex.Lock()
	defer mapper.decodedMutex.Unlock()
	delete(mapper.decodedMsgTkns, getHashValue(newTkn))

	return cmpTkn.CompressedToken, cmpTkn.Seq, nil
}

// func AddRespToken(newToken, compressedToken []byte, seq uint8) error {
// 	respTknTable[getHashValue(newToken)] = &compressedTknPair{CompressedToken: compressedToken, Seq: seq}
// 	return nil
// }

// func GetRespToken(token []byte) (*compressedTknPair, error) {
// 	tknPair, ok := respTknTable[getHashValue(token)]
// 	if !ok {
// 		return nil, errors.New("wrong token is sended")
// 	}

// 	return tknPair, nil
// }
