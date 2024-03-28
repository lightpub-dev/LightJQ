package internal

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
)

func encodeMsg(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg is nil")
	}
	return msgpack.Marshal(msg)
}

func decodeMsg(data []byte, msg interface{}) error {
	if data == nil {
		return fmt.Errorf("data is nil")
	}
	return msgpack.Unmarshal(data, msg)
}
