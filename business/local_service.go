package business

import (
	"strings"

	"github.com/AthenZ/athenz/libs/go/athenz-common/log"
	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/local_service"
)

type LocalServer struct {
	TcpServer       *local_service.TCPServer
	WebSocketServer *local_service.WebSocketServer
	Config          bean.Config
}

func NewLocalServer(config bean.Config, handleData func(clientAddr string, data []byte), handleConnected func(clientAddr string), handleDisconnected func(clientAddr string)) *LocalServer {
	localServer := &LocalServer{}

	tcpServer := local_service.NewTCPServer(handleData, handleConnected, handleDisconnected)
	go tcpServer.Start(config.TcpInfo.Port)

	webSocketServer := local_service.NewWebSocketServer(handleData, handleConnected, handleDisconnected)
	go webSocketServer.Start(config.WsInfo.Port, config.WsInfo.Path)

	localServer.TcpServer = tcpServer
	localServer.WebSocketServer = webSocketServer
	localServer.Config = config

	return localServer
}

func (localServer LocalServer) SendMessage(deviceData *DeviceData) {
	var iLocalServer local_service.ILocalServer
	switch strings.ToLower(deviceData.DeviceGateway.Protocol) {
	case "tcp":
		iLocalServer = localServer.TcpServer
	case "websocket":
		iLocalServer = localServer.WebSocketServer
	}
	err := iLocalServer.SendMessage(deviceData.DeviceGateway.Ip, deviceData.Data)
	if err != nil {
		log.Errorf("SendMessage message %+v", err)
	}
}

func (localServer LocalServer) IsOnline(ip string) bool {
	return localServer.TcpServer.IsOnline(ip)
}
