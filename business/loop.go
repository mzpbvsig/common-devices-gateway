package business

import (
    "time"
    log "github.com/sirupsen/logrus"

)

func makeDeviceDataTimeLoop() {
    for {
        for _, deviceGateway := range config.DeviceGateways {
            if deviceGateway.Interval == 0 {
                deviceGateway.Interval = 3
            }
            if !deviceGateway.IsOnline {
                log.Warnf("MakeDeviceDataTimeLoop deviceGateway %s is offline", deviceGateway.Ip)
                time.Sleep(time.Duration(deviceGateway.Interval) * time.Second)
                continue
            }
            for _, device := range deviceGateway.Devices {
                for _, entity := range device.Entities {
                    data, err := dp.ProcessMakeDeviceData(device, entity)
                    if err != nil {
                        log.Errorf("MakeDeviceDataTimeLoop error: %+v ", err)
                        continue
                    }
                    sendData := &DeviceData{
                        DeviceGateway: deviceGateway,            
                        Entity: entity,
                        Data:   data,
                        Type:   TimeLoop,
                    }
                    dataManager.Push(deviceGateway.Id, sendData)  
                    time.Sleep(time.Duration(deviceGateway.Interval) * time.Second)
                }
            }
        }
    }
}

func sendDeviceDataLoop() {
    for _, deviceGateway := range config.DeviceGateways {
        go sendDeviceDataLoopByGatewayId(deviceGateway.Id)
    }
}

func sendDeviceDataLoopByGatewayId(gatewayId string){
    for {   
		sendData := dataManager.Pop(gatewayId)
        if sendData!=nil && sendData.Entity != nil {
            log.Printf("Pop Data is %+v", sendData.Entity.EntityClass)
            dataManager.SetData(gatewayId, sendData)
            localServer.SendMessage(sendData)
            if restManager.isSearchDone(sendData) {
                mysqlManager.UpdateSearchDone(gatewayId)
                loadDevices(gatewayId)
                restManager.reset(gatewayId)
                log.Printf("Search Done")
            }
            time.Sleep(100 * time.Millisecond)
        }
        select {
			case <-nextSendDataChan:
			case <-time.After(time.Second):
			case <-stopChan:
				return 
        }
    }
}