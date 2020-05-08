package protocol

import (
	"encoding/json"
)

type Protocol struct {
	Op  byte   `json:"op"`
	Key string `json:"key"`
	Val string `json:"val,omitempty"`
	Exp string `json:"exp,omitempty"`
}

func (p *Protocol) GetJsonBytes() ([]byte, error) {
	jsonBytes, err := json.Marshal(*p)
	return jsonBytes, err
}

func GetDatagram(data []byte) (*Protocol, error) {
	var p Protocol
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
