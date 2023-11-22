package main

import (
	"testing"
	"github.com/mzpbvsig/common-devices-gateway/protocol"
)

func TestMakeData(t *testing.T) {
	address := byte(3)
	funcCode := byte(3)
	registerAddress := uint16(0x01F4)
	dataLength := uint16(0x0002)

	expected := []byte{0x03, 0x03, 0x01, 0xF4, 0x00, 0x02, 0x85, 0xE7}

	modbus := &protocol.Modbus{}
	result := modbus.MakeData(address, funcCode, registerAddress, dataLength)

	if !bytesEqual(result, expected) {
		t.Errorf("Generated data does not match the expected data, expected: %v, actual: %v", expected, result)
	}
}

func TestParseModbusResponse(t *testing.T) {
	// Simulate Modbus response data
	modbusResponse := []byte{0x03, 0x03, 0x04, 0x02, 0x09, 0xFF, 0x9B, 0x79, 0xFD}

	// Call the parseModbusResponse function to parse the Modbus response data
	modbus := &protocol.Modbus{}
	err := modbus.ParseResponse(modbusResponse)

	// Check if there's an error
	if err != nil {
		t.Errorf("Failed to parse Modbus response: %v", err)
		return
	}

	// Check if the parsed data is correct
	if modbus.Address != 0x03 {
		t.Errorf("Slave address does not match, expected: 0x01, actual: 0x%X", modbus.Address)
	}

	if modbus.FuncCode != 0x03 {
		t.Errorf("Function code does not match, expected: 0x03, actual: 0x%X", modbus.FuncCode)
	}

	if modbus.DataLength != 0x04 {
		t.Errorf("Data length does not match, expected: 0x04, actual: 0x%X", modbus.DataLength)
	}

	expectedData := []byte{0x02, 0x09, 0xFF, 0x9B}
	if !bytesEqual(modbus.Data, expectedData) {
		t.Errorf("Data does not match, expected: %v, actual: %v", expectedData, modbus.Data)
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
