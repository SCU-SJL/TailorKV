package protocol

import (
	"TailorKV/src/tailor"
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

type KeysDatagram struct {
	Keys []string `json:"keys"`
}

func (k *KeysDatagram) GetKeysJson(kvs []tailor.KV) ([]byte, error) {
	for _, kv := range kvs {
		k.Keys = append(k.Keys, kv.Key())
	}
	jsonBytes, err := json.Marshal(k)
	return jsonBytes, err
}

func GetKeys(data []byte) ([]string, error) {
	var k KeysDatagram
	err := json.Unmarshal(data, &k)
	if err != nil {
		return nil, err
	}
	return k.Keys, nil
}
