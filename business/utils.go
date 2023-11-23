package business

import (
	"fmt"

	"github.com/mzpbvsig/common-devices-gateway/utils"
)

func mergeIfLongEnough(base, toMerge string) string {
	if len(toMerge) > 8 {
		mergedData, err := utils.MergeJSONStrings(base, toMerge)
		if err == nil {
			return mergedData
		}
	}
	return base
}

func dispatchEvent(entityId string, eventMethod string, value string, Type DeviceType) error {
	deviceData := dataManager.GetQuickDeviceData(entityId)

	if deviceData == nil {
		return fmt.Errorf("device data not found by entity id %s", entityId)
	}

	deviceGateway := deviceData.DeviceGateway
	if !deviceGateway.IsOnline {
		return fmt.Errorf("device gateway id:%s, ip: %s is offline", deviceGateway.Id, deviceGateway.Ip)
	}

	entity := deviceData.Entity
	device := deviceData.Device

	for _, event := range entity.EntityClass.Events {
		if eventMethod == event.Method {
			if len(event.Code) > 7 {
				entity.EntityClass.Code = event.Code
			}
			entity.EntityClass.Data = mergeIfLongEnough(entity.EntityClass.Data, event.Data)
			break
		}
	}

	entity.EntityClass.Data = mergeIfLongEnough(entity.EntityClass.Data, value)
	entity.EntityClass.Data = mergeIfLongEnough(entity.EntityClass.Data, entity.Data)

	data, err := dp.ProcessRequest(deviceData.Device, deviceData.Entity)
	if err != nil {
		return err
	}

	sendData := &DeviceData{
		DeviceGateway: deviceGateway,
		Data:          data,
		Entity:        entity,
		Device:        device,
		Type:          Type,
	}
	dataManager.Unshift(deviceGateway.Id, sendData)

	return nil
}
