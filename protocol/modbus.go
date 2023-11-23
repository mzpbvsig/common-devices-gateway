package protocol

import (
	"bytes"
	"fmt"

	"github.com/mzpbvsig/common-devices-gateway/utils"
)

type Modbus struct {
	Request ModbusRequest
}

type ModbusRequest struct {
	Address         []byte
	FuncCode        byte
	RegisterAddress uint16
	RegisterLength  *uint16
	ByteCount       *byte
	Value           []uint16
}

func (modbus *Modbus) MakeData(request *ModbusRequest) []byte {
	data := append(request.Address, request.FuncCode)

	data = append(data, uint8(request.RegisterAddress>>8), uint8(request.RegisterAddress))

	if request.RegisterLength != nil {
		data = append(data, uint8(*request.RegisterLength>>8), uint8(*request.RegisterLength))
	}

	if request.ByteCount != nil {
		data = append(data, *request.ByteCount)
	}

	if request.Value != nil && len(request.Value) > 0 {
		for _, val := range request.Value {
			highByte := byte(val >> 8)
			lowByte := byte(val & 0xFF)
			data = append(data, highByte, lowByte)
		}
	}

	crc := utils.CalculateCRC(data)

	data = append(data, uint8(crc), uint8(crc>>8))

	return data
}

func (modbus *Modbus) ParseResponse(request *ModbusRequest, response []byte) ([]byte, error) {
	// 多字节地址支持
	addressLength := 1
	if request.Address != nil {
		addressLength = len(request.Address)
	}

	// 地址长度 + 功能码长度 + CRC长度
	if len(response) < addressLength+3 {
		return nil, fmt.Errorf("response data length is insufficient")
	}

	// 提取并构造接收到的CRC
	receivedCRCHigh := uint8(response[len(response)-2])
	receivedCRCLow := uint8(response[len(response)-1])

	// 计算CRC
	calculatedCRC := utils.CalculateCRC(response[:len(response)-2])

	// 校验CRC
	if uint8(calculatedCRC) != receivedCRCHigh || uint8(calculatedCRC>>8) != receivedCRCLow {
		return nil, fmt.Errorf("crc check failed")
	}

	// 解析地址码和功能码
	address := response[0:addressLength]
	funcCode := response[addressLength]
	if funcCode != request.FuncCode || !bytes.Equal(request.Address, address) {
		return nil, fmt.Errorf("address or funcode is not the same")
	}

	return response, nil
}
