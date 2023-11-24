package local_service

import (
	"sync"
)

type ILocalServer interface {
	Stop()
	SendMessage(ip string, data []byte) error
}

type BaseServer struct {
	clientsLock          sync.Mutex
	shutdownChan         chan struct{}
	MessageCallback      func(ip string, data []byte)
	ConnectedCallback    func(ip string)
	DisconnectedCallback func(ip string)
}
