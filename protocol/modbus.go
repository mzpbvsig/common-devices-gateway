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
	RegisterLength  uint16
	Value           []uint16
}

func (modbus *Modbus) MakeData(request *ModbusRequest) []byte {
	// 初始化请求数据，包括设备地址和功能码
	data := append(request.Address, request.FuncCode)

	// 添加寄存器地址。由于寄存器地址可能大于255，需要分为高位和低位
	data = append(data, uint8(request.RegisterAddress>>8), uint8(request.RegisterAddress))

	// 添加寄存器长度。同样，寄存器长度可能大于255，需要分为高位和低位
	data = append(data, uint8(request.RegisterLength>>8), uint8(request.RegisterLength))

	// 寄存器写入内容
	if request.Value != nil && len(request.Value) > 0 {
		for _, val := range request.Value {
			highByte := byte(val >> 8)
			lowByte := byte(val & 0xFF)
			data = append(data, highByte, lowByte)
		}
	}

	// 计算请求数据的 CRC 校验码
	crc := utils.CalculateCRC(data)

	// 将 CRC 校验码添加到请求数据的末尾。CRC 是一个16位数，因此分为高位和低位
	data = append(data, uint8(crc), uint8(crc>>8))

	// 返回构建好的请求数据
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
