package devices

import (
	"github.com/mzpbvsig/common-devices-gateway/bean"
    "github.com/mzpbvsig/common-devices-gateway/utils"
    "github.com/mzpbvsig/common-devices-gateway/devices/modbus"

    "fmt"
)


type DataFromDeviceProcessor func(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error)
type MakeDeviceDataProcessor func(device *bean.Device, entity *bean.Entity) ([]byte,error)


const (
    Modbus = "modbus"
    ModbusPlus = "modbus_plus"
)

type DeviceProcessor struct {
    protocols map[string]Protocol

    dataFromDeviceProcessor map[string]DataFromDeviceProcessor
    makeDeviceDataProcessor map[string]MakeDeviceDataProcessor
}

func NewDeviceProcessor() *DeviceProcessor {
    dp := &DeviceProcessor{
        protocols: make(map[string]Protocol),
        dataFromDeviceProcessor: make(map[string]DataFromDeviceProcessor),
        makeDeviceDataProcessor: make(map[string]MakeDeviceDataProcessor),
    }

    dp.protocols[Modbus] = modbus.NewModbus(func(sn int) []byte {
        return []byte{byte(sn)}
    })

    dp.protocols[ModbusPlus] = modbus.NewModbus(func(sn int) []byte {
        address,_ := utils.IntToBytes(sn)
        return address
    })


    //Register processors for different device models
    dp.registerDataFromDeviceProcessors()
    dp.registerMakeDeviceDataProcessors()
    
    return dp
}

func (dp *DeviceProcessor)  LoadPulgins(path string){

}

func (dp *DeviceProcessor)  LoadPulgin(name, path string){

}

func (dp *DeviceProcessor)  UnloadPulgin(name string){

}


// 根据设备协议构造不同的处理方法
func (dp *DeviceProcessor) registerDataFromDeviceProcessors() {
    dp.dataFromDeviceProcessor[Modbus] = dp.protocols[Modbus].ProcessDataFromDevice
    dp.dataFromDeviceProcessor[ModbusPlus] = dp.protocols[ModbusPlus].ProcessDataFromDevice
}

func (dp *DeviceProcessor) registerMakeDeviceDataProcessors() {
    dp.makeDeviceDataProcessor[Modbus] = dp.protocols[Modbus].MakeDeviceData
    dp.makeDeviceDataProcessor[ModbusPlus] = dp.protocols[ModbusPlus].MakeDeviceData
}

// 处理设备返回的数据
func (dp *DeviceProcessor) ProcessDataFromDevice(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool)  (string,error) {
    processor, ok := dp.dataFromDeviceProcessor[device.DeviceClass.Protocol]
    if !ok {
        return "",fmt.Errorf("Device model %s is not supported ProcessDataFromDevice", device.DeviceClass.Protocol)
    }

    return processor(device, entity, data, isRunJs)
}

// 处理构造设备的数据
func (dp *DeviceProcessor) ProcessMakeDeviceData(device *bean.Device, entity *bean.Entity) ([]byte, error) {
    processor, ok := dp.makeDeviceDataProcessor[device.DeviceClass.Protocol]
    if !ok {
        return nil, fmt.Errorf("Device model %s is not supported ProcessMakeDeviceData", device.DeviceClass.Protocol)
    }
    return processor(device, entity)
}