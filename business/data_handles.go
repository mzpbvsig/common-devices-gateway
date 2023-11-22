package business

import (
	log "github.com/sirupsen/logrus"
	"encoding/json"

    "github.com/mzpbvsig/common-devices-gateway/bean"


)

type DataHandler interface {
    Handle(deviceGateway *bean.DeviceGateway, deviceData *DeviceData, data []byte) bool
}

type CommonHandler struct{}
func (h *CommonHandler) Handle(deviceGateway *bean.DeviceGateway, deviceData *DeviceData, data []byte) bool {

    device := deviceData.Device
    entity := deviceData.Entity
	result, err := dp.ProcessDataFromDevice(device, entity, data, true)

    if err != nil {
        log.Errorf("Error processing data from device: %+v", err)
        return false
    }

    entity.State = result    
    jsonData, err := json.Marshal(entity)
    if err != nil {
        log.Errorf("Json.Marshal error %+v", err)
        return  false
    }	
    cloudServer.ReportState(device.DeviceClass.Type, deviceGateway.Id, jsonData)

    log.Printf("Handle Data continue next data common")

	return true
}

type TestHandler struct{
	  
}
func (h *TestHandler) Handle(deviceGateway *bean.DeviceGateway, deviceData *DeviceData, data []byte) bool {

    device := deviceData.Device
    entity := deviceData.Entity
	result, err := dp.ProcessDataFromDevice(device, entity, data, true)
    if err != nil {
        log.Error("Error processing data from device: ", err)
        return false
    }

    if restManager.TestClassback != nil {
        entity.State =  result
         restManager.TestClassback(entity)
    }

    return true
}

type SearchHandler struct{}
func (h *SearchHandler) Handle(deviceGateway *bean.DeviceGateway, deviceData *DeviceData, data []byte) bool {
    device := deviceData.Device
    entity := deviceData.Entity

    _, err := dp.ProcessDataFromDevice(device, entity, data, false)
    if err != nil {
        log.Error("Error processing search data from device: ", err)
        return false
    } else {
		restManager.searched(device, entity)
    }

    log.Printf("Handle Data continue next data from search")

	return true
}


// 将处理器映射到相应的设备类型
var dataHandlers = map[DeviceType]DataHandler{
    TimeLoop: &CommonHandler{},
    Search: &SearchHandler{},
    Test: &TestHandler{},
    Cloud: &CommonHandler{}, 
}