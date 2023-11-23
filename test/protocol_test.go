package main

import (
	"testing"

	"github.com/mzpbvsig/common-devices-gateway/protocol"
)

func TestMakeData(t *testing.T) {
	registerLength := uint16(2)
	request := &protocol.ModbusRequest{
		Address:         []byte{3},
		FuncCode:        3,
		RegisterAddress: 0x01F4,
		RegisterLength:  &registerLength,
	}

	expected := []byte{0x03, 0x03, 0x01, 0xF4, 0x00, 0x02, 0x85, 0xE7}

	modbus := &protocol.Modbus{}
	result := modbus.MakeData(request)

	if !bytesEqual(result, expected) {
		t.Errorf("Generated data does not match the expected data, expected: %v, actual: %v", expected, result)
	}
}

func TestParseModbusResponse(t *testing.T) {
	data := []byte{0x03, 0x03, 0x01, 0xF4, 0x00, 0x02, 0x85, 0xE7}

	request := &protocol.ModbusRequest{
		Address:  []byte{3},
		FuncCode: 3,
	}
	modbus := &protocol.Modbus{}
	_, err := modbus.ParseResponse(request, data)

	if err != nil {
		t.Errorf("Failed to parse Modbus response: %v", err)
		return
	}

}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
