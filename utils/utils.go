package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
)

func CalculateCRC(data []byte) uint16 {
	crc := uint16(0xFFFF)
	polynomial := uint16(0xA001)

	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			lsb := crc & 0x0001
			crc >>= 1
			if lsb == 1 {
				crc ^= polynomial
			}
		}
	}

	return crc
}

func IntToBytes(sn int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int32(sn))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func BytesToInt(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, errors.New("byte slice is too short")
	}
	var intValue int32
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, &intValue)
	if err != nil {
		return 0, err
	}
	return int(intValue), nil
}

func MergeJSONStrings(jsonStr1, jsonStr2 string) (string, error) {
	var map1, map2 map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr1), &map1)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(jsonStr2), &map2)
	if err != nil {
		return "", err
	}

	for k, v := range map2 {
		map1[k] = v
	}

	mergedJSON, err := json.Marshal(map1)
	if err != nil {
		return "", err
	}

	return string(mergedJSON), nil
}
