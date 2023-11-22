package devices

import (
	"github.com/mzpbvsig/common-devices-gateway/bean"
)

type Protocol interface {
    ProcessDataFromDevice(device *bean.Device, entity *bean.Entity, data []byte, isRunJs bool) (string, error)
	MakeDeviceData(device *bean.Device, entity *bean.Entity) ([]byte, error)
}