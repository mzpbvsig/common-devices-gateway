package local_service

type Server interface {
	Start(port int)
	Stop()
	SendMessage(clientAddr string, data []byte) error
}
