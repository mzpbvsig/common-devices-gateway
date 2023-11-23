package business

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/cloud_service"

	log "github.com/sirupsen/logrus"
)

type CloudServer struct {
	PulsarManager *cloud_service.PulsarManager
	Config        bean.Config
}

func NewCloudServer(config bean.Config) *CloudServer {
	cloudServer = &CloudServer{}

	pulsarManager := cloud_service.NewPulsarManager(config.PulsarServer)
	pulsarManager.CreateRegisterProducer()

	cloudServer.PulsarManager = pulsarManager
	cloudServer.Config = config

	cloudServer.CreateStateProducers()
	pulsarManager.ListenForEvents(cloudServer.handlePulsarEvent)

	return cloudServer
}

func (cloudServer CloudServer) handlePulsarEvent(payload []byte) {
	log.Printf("Receviced cloud message: %s", string(payload))
	var eventMap map[string]interface{}
	err := json.Unmarshal(payload, &eventMap)
	if err != nil {
		return
	}

	entityId, ok := eventMap["entity_id"].(string)
	eventMethod, _ := eventMap["event_method"].(string)
	data, _ := eventMap["data"].(string)
	if !ok {
		log.Errorf("Cloud message error extracting entity_id")
		return
	}

	err = dispatchEvent(entityId, eventMethod, data, Cloud)
	if err != nil {
		log.Errorf("Cloud message make data error: %+v ", err)
	}
}

func (server CloudServer) CreateStateProducers() {
	server.PulsarManager.CreateStateProducers(config.DeviceGateways)
}

func (server CloudServer) Register(deviceGateway *bean.DeviceGateway) {
	ctx := context.Background()
	data, err := json.Marshal(deviceGateway)
	if err != nil {
		log.Errorf("Register JSON encoding failed: %s", err)
		return
	}
	switch server.Config.CloudProtocol {
	case "pulsar":
		server.PulsarManager.Register(ctx, data)
	}
}

func (server CloudServer) Registers() {
	for _, deviceGateway := range config.DeviceGateways {
		server.Register(deviceGateway)
		time.Sleep(100 * time.Millisecond)
	}
}

func (server CloudServer) RegisterByGatewayId(gatewayId string) {
	if len(gatewayId) == 0 || gatewayId == "0" {
		server.Registers()
	} else {
		deviceGateway := getDeviceGatewayById(gatewayId)
		if deviceGateway != nil {
			server.Register(deviceGateway)
		}
	}
}

func (server CloudServer) ReportState(deviceType string, gatewayId string, data []byte) {
	ctx := context.Background()
	switch server.Config.CloudProtocol {
	case "pulsar":
		server.PulsarManager.ReportState(ctx, gatewayId, data)
	}
}
