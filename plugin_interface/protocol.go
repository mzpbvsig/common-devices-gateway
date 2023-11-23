package plugin_interface

import (
	"fmt"
	"plugin"

	"github.com/mzpbvsig/common-devices-gateway/bean"
)

type Protocol interface {
	ProcessDataFromDevice(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error)
	MakeDeviceData(device *bean.Device, entity *bean.Entity) ([]byte, error)
}

func GetProtocolInstance(pluginPath string) (Protocol, error) {
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}

	sym, err := plug.Lookup("NewPluginInstance")
	if err != nil {
		return nil, err
	}

	newFunc, ok := sym.(func() Protocol)
	if !ok {
		return nil, fmt.Errorf("plugin function has incorrect type")
	}

	return newFunc(), nil
}
