package modbus

import (
	"github.com/mzpbvsig/common-devices-gateway/protocol"
	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/code_engin"
    "encoding/json"
	"strconv"
    "fmt"
)

type Modbus struct {
	addressProvider func(int) []byte
	globalEngine  *code_engin.JSEngine
	modbus *protocol.Modbus
}

type ModbusData struct {
	RegisterAddress string   `json:"register_address"`
	FuncCode string   `json:"func_code"`
	RegisterLength string `json:"register_length"`
	Data    string `json:"data"`
	Params   string `json:"params"`
}

func NewModbus(addressProvider func(int) []byte) *Modbus {
    return &Modbus{
        addressProvider: addressProvider,
		modbus: &protocol.Modbus{},
		globalEngine: code_engin.NewJSEngine(),
    }
}

func (dModbus *Modbus) ProcessDataFromDevice(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error) {
	sn, err := strconv.Atoi(device.SN)
	if err != nil {
		return "Device sn convert error", err
	}
	address := dModbus.addressProvider(sn)

	var modbusData ModbusData
	err = json.Unmarshal([]byte(entity.EntityClass.Data), &modbusData)
	if err != nil {
		return "ModbusData parse error", err
	}

	request, err := dModbus.convertToRequest(modbusData)
	if err != nil {
		return "Convert ModbusData To Request", err
	}

	request.Address = address
	response, err := dModbus.modbus.ParseResponse(request, data)
    if err != nil {
        return "", err
    }
	
	if isRunJs {
		jsCode := entity.EntityClass.Code
	    if len(modbusData.Params) > 6 {
			jsCode = fmt.Sprintf(`var paramsObj = JSON.parse('%s');%s`, modbusData.Params, jsCode)
		}
		result, err := dModbus.globalEngine.RunJs(jsCode, response)
		if err != nil {
			return fmt.Sprintf("%+v", err), nil
		} else {
			return result, nil
		}
	}
	
	return "", nil
}

func (dModbus *Modbus) MakeDeviceData(device *bean.Device, entity *bean.Entity) ([]byte, error) {
	return dModbus.makeModbusData(device, entity)
}


func (dModbus *Modbus) makeModbusData(device *bean.Device, entity *bean.Entity) ([]byte, error) {
	sn, err := strconv.Atoi(device.SN)
	if err != nil {
		return nil, err
	}

	address := dModbus.addressProvider(sn)

	var modbusData ModbusData
	err = json.Unmarshal([]byte(entity.EntityClass.Data), &modbusData)
	if err != nil {
		return nil, err
	}
    
	request, err := dModbus.convertToRequest(modbusData)
	if err != nil {
		return nil, err
	}
	request.Address = address;

	return dModbus.modbus.MakeData(request), nil

}

func (dModbus *Modbus) convertToRequest(data ModbusData) (*protocol.ModbusRequest, error) {
    request := &protocol.ModbusRequest{}

    // 转换 RegisterAddress
    if data.RegisterAddress != "" {
        registerAddress, err := strconv.ParseUint(data.RegisterAddress, 10, 16)
        if err != nil {
            return nil, fmt.Errorf("Error converting RegisterAddress to byte: %+v", err)
        }
        request.RegisterAddress = uint16(registerAddress)
    } else {
		return nil, fmt.Errorf("RegisterAddress is empty.")
	}

    // 转换 FuncCode
    if data.FuncCode != "" {
        funcCode, err := strconv.ParseUint(data.FuncCode, 10, 8)
        if err != nil {
            return nil, fmt.Errorf("Error converting FuncCode to byte: %+v", err)
        } else {
			request.FuncCode = byte(funcCode)
		}
    } else {
		return nil, fmt.Errorf("FuncCode is empty.")
	}

    // 转换 RegisterLength
    if data.RegisterLength != "" {
        registerLength, err := strconv.ParseUint(data.RegisterLength, 10, 16)
        if err != nil {
            return nil, fmt.Errorf("Error converting RegisterLength to byte: %+v", err)
        } else {
			request.RegisterLength = uint16(registerLength)
		}
    } else {
		return nil, fmt.Errorf("RegisterLength is empty.")
	}

	// 转换 Value
	if data.Data != "" {
		var value []uint16
		err := json.Unmarshal([]byte(data.Data), &value)
		if err != nil {
			request.Value = nil
		} else {
			request.Value = value
		}
	}

    return request, nil
}
