package business

import (
    "github.com/mzpbvsig/common-devices-gateway/bean"
    log "github.com/sirupsen/logrus"
	"strings"

)

func loadConfig(){
	config = bean.GetConfig[bean.Config]("conf.yaml")
}


func loadDeviceClasses(){
	deviceClasses, err := mysqlManager.LoadDeviceClasses()
	if err != nil {
		log.Errorf("Load device classes error :%s", err)
		return 
	}
	config.DeviceClasses = deviceClasses
	log.Printf("Load device classes :%+v", deviceClasses)
}

func loadDeviceGateways(){
	deviceGateways, err := mysqlManager.LoadDeviceGateways()
	if err != nil {
		log.Errorf("Load device gateways error :%s", err)
		return
	}
	updateDeviceGateways(deviceGateways)
	log.Printf("Load device gateways :%+v", deviceGateways)
}

func updateDeviceGateways(deviceGateways []*bean.DeviceGateway){
	for _, deviceGateway := range config.DeviceGateways {
		for _, cdeviceGateway := range deviceGateways {	
			if deviceGateway.Id == cdeviceGateway.Id {
				cdeviceGateway.IsOnline = deviceGateway.IsOnline
				break
			}
		}
	}
	config.DeviceGateways = deviceGateways
}

func getDeviceGateway(clientAddr string) *bean.DeviceGateway {
    for _, deviceGateway := range config.DeviceGateways {
        if strings.Index(clientAddr, deviceGateway.Ip) != -1 {
             return deviceGateway
        }
    }
    return nil
}

func loadDevices(gatewayId string){
	for _, deviceGateway := range config.DeviceGateways {
		if gatewayId == deviceGateway.Id {
			devices, err := mysqlManager.GetDevices(gatewayId)
			if err != nil {
				log.Errorf("LoadDevices from gatewayId %+v", err)
				break
			}
			deviceGateway.Devices = devices
			break
		}
	}
}