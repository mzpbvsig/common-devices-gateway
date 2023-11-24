package local_service

import (
	"sync"
)

type ILocalServer interface {
	Stop()
	SendMessage(clientAddr string, data []byte) error
}

type BaseServer struct {
	clientsLock          sync.Mutex
	shutdownChan         chan struct{}
	MessageCallback      func(clientAddr string, data []byte)
	ConnectedCallback    func(clientAddr string)
	DisconnectedCallback func(clientAddr string)
}
