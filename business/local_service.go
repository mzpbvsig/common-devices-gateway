package business

import (
	"strings"

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
	switch strings.ToLower(deviceData.DeviceGateway.Protocol) {
	case "tcp":
		localServer.TcpServer.SendMessage(deviceData.DeviceGateway.Ip, deviceData.Data)
	case "websocket":
		localServer.WebSocketServer.SendMessage(deviceData.DeviceGateway.Ip, deviceData.Data)
	}
}
