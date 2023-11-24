package business

import (
	log "github.com/sirupsen/logrus"
)

func handleData(clientAddr string, data []byte) {
	deviceGateway := getDeviceGatewayByClientAddr(clientAddr)
	if deviceGateway == nil {
		log.Error("Devcie gateway not found: %s" + clientAddr)
		return
	}

	deviceData := dataManager.GetData(deviceGateway.Id)
	if deviceData == nil || deviceData.Entity == nil {
		nextSendDataChan <- true
		log.Error("Device data not found from device gateway " + deviceGateway.Id)
		return
	}

	handler, exists := dataHandlers[deviceData.Type]
	if exists {
		if handler.Handle(deviceGateway, deviceData, data) {
			dataManager.Clear(deviceGateway.Id)
			nextSendDataChan <- true
		}
	}
}

func handleConnected(clientAddr string) {
	deviceGateway := getDeviceGatewayByClientAddr(clientAddr)
	if deviceGateway != nil {
		deviceGateway.IsOnline = true
	}
}

func handleDisconnected(clientAddr string) {
	deviceGateway := getDeviceGatewayByClientAddr(clientAddr)
	if deviceGateway != nil {
		deviceGateway.IsOnline = false
	}
}
