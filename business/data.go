package business

import (
	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/data_struct"
)

type DeviceType int

const (
	TimeLoop DeviceType = iota
	Search
	Test
	Cloud
)

type DeviceData struct {
	DeviceGateway *bean.DeviceGateway
	Entity        *bean.Entity
	Device        *bean.Device
	Type          DeviceType
	Data          []byte
}

type DataManager struct {
	Queues      map[string]*data_struct.Queue[*DeviceData]
	DeviceDatas map[string]*DeviceData

	QuickDeviceDatas map[string]*DeviceData
}

// NewDataManager creates a new instance of DataManager with an initialized queue.
func NewDataManager() *DataManager {
	dataManager := &DataManager{
		Queues:      make(map[string]*data_struct.Queue[*DeviceData]),
		DeviceDatas: make(map[string]*DeviceData),

		QuickDeviceDatas: make(map[string]*DeviceData),
	}
	dataManager.BuildQuickDeviceDatas()
	return dataManager
}

func (manager *DataManager) BuildQuickDeviceDatas() {
	for _, deviceGateway := range config.DeviceGateways {
		for _, device := range deviceGateway.Devices {
			for _, entity := range device.Entities {
				manager.QuickDeviceDatas[entity.Id] = &DeviceData{
					DeviceGateway: deviceGateway,
					Device:        device,
					Entity:        entity,
				}
			}
		}
	}
}

func (manager *DataManager) GetQuickDeviceData(entityId string) *DeviceData {
	return manager.QuickDeviceDatas[entityId]
}

func (manager *DataManager) SetData(gatewayId string, deviceData *DeviceData) {
	manager.DeviceDatas[gatewayId] = deviceData
}

func (manager *DataManager) Push(gatewayId string, data *DeviceData) {
	if manager.Queues[gatewayId] == nil {
		manager.Queues[gatewayId] = data_struct.NewQueue[*DeviceData]()
	}

	manager.Queues[gatewayId].Push(data)
}

func (manager *DataManager) Unshift(gatewayId string, data *DeviceData) {
	if manager.Queues[gatewayId] == nil {
		manager.Queues[gatewayId] = data_struct.NewQueue[*DeviceData]()
	}

	manager.Queues[gatewayId].Unshift(data)
}

func (manager *DataManager) Pop(gatewayId string) *DeviceData {
	if manager.Queues[gatewayId] == nil {
		return nil
	}

	return manager.Queues[gatewayId].Pop()
}

func (manager *DataManager) Clear(gatewayId string) {
	deviceData, exists := manager.DeviceDatas[gatewayId]
	if exists {
		deviceData.DeviceGateway = nil
		deviceData.Entity = nil
		deviceData.Device = nil
		manager.DeviceDatas[gatewayId] = nil
	}
}

func (manager *DataManager) GetData(gatewayId string) *DeviceData {
	return manager.DeviceDatas[gatewayId]
}

func (manager *DataManager) RemoveAll(gatewayId string) {
	if manager.Queues[gatewayId] == nil {
		manager.Queues[gatewayId] = data_struct.NewQueue[*DeviceData]()
	}

	manager.Queues[gatewayId].RemoveAll(func(item *DeviceData) bool {
		if item.Type != Search {
			return false
		}

		if item.Device == nil {
			return false
		}

		return item.Device.GatewayId == gatewayId
	})
}
