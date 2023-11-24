package devices

import (
	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/code_engin"

	"fmt"
)

type DeviceProcessor struct {
	protocols     map[string]*bean.Protocol
	globalEngines map[string]*code_engin.JSEngine
}

func NewDeviceProcessor(protocols []*bean.Protocol) *DeviceProcessor {
	dp := &DeviceProcessor{
		protocols:     make(map[string]*bean.Protocol),
		globalEngines: make(map[string]*code_engin.JSEngine),
	}

	dp.LoadProtocols(protocols)

	return dp
}

func (dp *DeviceProcessor) LoadProtocols(protocols []*bean.Protocol) {
	for _, protocol := range protocols {
		dp.protocols[protocol.Name] = protocol
	}
}

func (dp *DeviceProcessor) ReloadProtocols(protocols []*bean.Protocol) {
	dp.LoadProtocols(protocols)
}

func (dp *DeviceProcessor) ProcessRequest(device *bean.Device, entity *bean.Entity) ([]byte, error) {
	protocol, ok := dp.protocols[device.DeviceClass.Protocol]
	if !ok {
		return nil, fmt.Errorf("device model %s is not supported request code", device.DeviceClass.Protocol)
	}
	if dp.globalEngines[device.GatewayId] == nil {
		dp.globalEngines[device.GatewayId] = code_engin.NewJSEngine()
	}
	return dp.globalEngines[device.GatewayId].Request(protocol.RequestCode, device, entity)
}

func (dp *DeviceProcessor) ProcessResponse(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error) {
	protocol, ok := dp.protocols[device.DeviceClass.Protocol]
	if !ok {
		return "", fmt.Errorf("device model %s is not supported response code", device.DeviceClass.Protocol)
	}
	if dp.globalEngines[device.GatewayId] == nil {
		dp.globalEngines[device.GatewayId] = code_engin.NewJSEngine()
	}
	return dp.globalEngines[device.GatewayId].Response(protocol.ResponseCode, device, entity, data)
}
