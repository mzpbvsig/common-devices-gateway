package bean

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
)

type Config struct {
	PulsarServer   PulsarServer
	CloudProtocol  string
	DeviceClasses  []DeviceClass
	DeviceGateways []*DeviceGateway
	TcpInfo        TcpInfo
	WsInfo         WsInfo
	MysqlOption    MysqlOption
	RestPort       int
	Protocols      []*Protocol
}

func GetConfig[T any](path string) T {
	var c T

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("Yaml file get err #%v ", err)
		return c
	}

	err = yaml.Unmarshal(yamlFile, &c)

	if err != nil {
		log.Errorf("Unmarshal: #%v", err)
	}

	return c
}
