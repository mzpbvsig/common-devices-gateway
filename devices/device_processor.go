package devices

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/devices/modbus"
	"github.com/mzpbvsig/common-devices-gateway/plugin_interface"
	"github.com/mzpbvsig/common-devices-gateway/utils"

	"fmt"

	log "github.com/sirupsen/logrus"
)

type DataFromDeviceProcessor func(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error)
type MakeDeviceDataProcessor func(device *bean.Device, entity *bean.Entity) ([]byte, error)

const (
	Modbus     = "modbus"
	ModbusPlus = "modbus_plus"
)

type DeviceProcessor struct {
	protocols map[string]plugin_interface.Protocol

	dataFromDeviceProcessor map[string]DataFromDeviceProcessor
	makeDeviceDataProcessor map[string]MakeDeviceDataProcessor
}

func NewDeviceProcessor() *DeviceProcessor {
	dp := &DeviceProcessor{
		protocols:               make(map[string]plugin_interface.Protocol),
		dataFromDeviceProcessor: make(map[string]DataFromDeviceProcessor),
		makeDeviceDataProcessor: make(map[string]MakeDeviceDataProcessor),
	}

	dp.protocols[Modbus] = modbus.NewModbus(func(sn int) []byte {
		return []byte{byte(sn)}
	})

	dp.protocols[ModbusPlus] = modbus.NewModbus(func(sn int) []byte {
		address, _ := utils.IntToBytes(sn)
		return address
	})

	//Register processors for different device models
	dp.registerDataFromDeviceProcessors()
	dp.registerMakeDeviceDataProcessors()

	return dp
}

func (dp *DeviceProcessor) LoadPlugins(directory string) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		log.Fatalf("Failed to read directory: %s", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".so") {
			pluginPath := filepath.Join(directory, entry.Name())
			pluginName := strings.TrimSuffix(entry.Name(), ".so")

			dp.LoadPlugin(pluginName, pluginPath)
		}
	}
}

func (dp *DeviceProcessor) LoadPlugin(name, path string) {
	protocol, err := plugin_interface.GetProtocolInstance(path)
	if err != nil {
		log.Error("Loaded plugin %s error: %+v path is %s", name, err, path)
		return
	}
	dp.protocols[name] = protocol
	log.Printf("Loaded plugin: %s", name)
}

func (dp *DeviceProcessor) UnloadPulgin(name string) {
	dp.protocols[name] = nil
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
func (dp *DeviceProcessor) ProcessDataFromDevice(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error) {
	processor, ok := dp.dataFromDeviceProcessor[device.DeviceClass.Protocol]
	if !ok {
		return "", fmt.Errorf("device model %s is not supported ProcessDataFromDevice", device.DeviceClass.Protocol)
	}

	return processor(device, entity, data, isRunJs)
}

// 处理构造设备的数据
func (dp *DeviceProcessor) ProcessMakeDeviceData(device *bean.Device, entity *bean.Entity) ([]byte, error) {
	processor, ok := dp.makeDeviceDataProcessor[device.DeviceClass.Protocol]
	if !ok {
		return nil, fmt.Errorf("device model %s is not supported ProcessMakeDeviceData", device.DeviceClass.Protocol)
	}
	return processor(device, entity)
}
