package devices

import (
	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/code_engin"

	"fmt"
)

type DeviceProcessor struct {
	protocols    map[string]*bean.Protocol
	globalEngine *code_engin.JSEngine
}

func NewDeviceProcessor(protocols []*bean.Protocol) *DeviceProcessor {
	dp := &DeviceProcessor{
		protocols:    make(map[string]*bean.Protocol),
		globalEngine: code_engin.NewJSEngine(),
	}
	for _, protocol := range protocols {
		dp.protocols[protocol.Name] = protocol
	}

	return dp
}

func (dp *DeviceProcessor) ProcessDataFromDevice(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error) {
	protocol, ok := dp.protocols[device.DeviceClass.Protocol]
	if !ok {
		return "", fmt.Errorf("device model %s is not supported ProcessDataFromDevice", device.DeviceClass.Protocol)
	}

	return dp.globalEngine.Response(protocol.ResponseCode, device, entity, data)
}

func (dp *DeviceProcessor) ProcessMakeDeviceData(device *bean.Device, entity *bean.Entity) ([]byte, error) {
	protocol, ok := dp.protocols[device.DeviceClass.Protocol]
	if !ok {
		return nil, fmt.Errorf("device model %s is not supported ProcessDataFromDevice", device.DeviceClass.Protocol)
	}

	return dp.globalEngine.Request(protocol.RequestCode, device, entity)
}
