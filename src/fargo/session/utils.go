package session

import (
	"bytes"
	"encoding/gob"
)

// encodeGob 生成 gob.
func encodeGob(obj map[interface{}]interface{}) (encode []byte, err error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(obj)
	if err != nil {
		return
	}
	encode = buf.Bytes()

	return
}

// decodeGob 解 gob.
func decodeGob(encoded []byte) (decode map[interface{}]interface{}, err error) {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&decode)
	if err != nil {
		return
	}

	return
}
