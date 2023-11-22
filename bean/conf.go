package bean

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
)



type Config struct {
	PulsarServer PulsarServer 
	CloudProtocol  string
	DeviceClasses  []DeviceClass
	DeviceGateways []*DeviceGateway
	TcpInfo TcpInfo
	WsInfo WsInfo
	MysqlOption MysqlOption
	RestPort    int
	IsOpenTimeLoop bool
}

func GetConfig[T any](path string) T {
	var c T

	yamlFile, err := ioutil.ReadFile(path)
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
