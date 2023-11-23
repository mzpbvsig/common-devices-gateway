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
	// 加载插件
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}

	// 查找 NewPluginInstance 函数
	sym, err := plug.Lookup("NewPluginInstance")
	if err != nil {
		return nil, err
	}

	// 断言函数的类型
	newFunc, ok := sym.(func() Protocol)
	if !ok {
		return nil, fmt.Errorf("plugin function has incorrect type")
	}

	// 调用函数获取实例
	return newFunc(), nil
}
