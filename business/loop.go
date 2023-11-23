package business

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func sendDeviceDataLoop() {
	for _, deviceGateway := range config.DeviceGateways {
		go sendDeviceDataLoopByGatewayId(deviceGateway.Id)
	}
}

func sendDeviceDataLoopByGatewayId(gatewayId string) {
	for {
		sendData := dataManager.Pop(gatewayId)
		if sendData != nil && sendData.Entity != nil {
			log.Printf("Pop data entity id is %s", sendData.Entity.Id)
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
