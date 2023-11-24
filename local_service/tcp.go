package local_service

import (
	"fmt"
	"net"

	"github.com/mzpbvsig/common-devices-gateway/utils"
	log "github.com/sirupsen/logrus"
)

// TCPServer is a TCP server object
type TCPServer struct {
	BaseServer
	listener net.Listener
	clients  map[string]net.Conn
}

// NewTCPServer creates a new TCPServer object
func NewTCPServer(messageCallback func(clientAddr string, data []byte), connectedCallback func(clientAddr string), disconnectedCallback func(clientAddr string)) *TCPServer {
	return &TCPServer{
		clients: make(map[string]net.Conn),
		BaseServer: BaseServer{
			shutdownChan:         make(chan struct{}),
			MessageCallback:      messageCallback,
			ConnectedCallback:    connectedCallback,
			DisconnectedCallback: disconnectedCallback,
		},
	}
}

func (s *TCPServer) Start(port int) {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Println("Failed to start the server:", err)
		return
	}
	defer s.listener.Close()

	log.Println("Server has started, waiting for client connections...")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Client connection error:", err)
			continue
		}

		go s.handleClient(conn)
	}
}

// Stop stops the TCP server
func (s *TCPServer) Stop() {
	close(s.shutdownChan)
	s.listener.Close()
}

func (s *TCPServer) IsOnline(ip string) bool {
	return s.clients[ip] != nil
}

func (s *TCPServer) handleClient(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	log.Printf("Client %s has connected", clientAddr)

	ip := utils.ExtractIP(clientAddr)

	s.clientsLock.Lock()
	s.clients[ip] = conn
	s.clientsLock.Unlock()

	if s.ConnectedCallback != nil {
		s.ConnectedCallback(ip)
	}

	for {
		select {
		case <-s.shutdownChan:
			log.Printf("Server is shutting down, client %s has disconnected", clientAddr)
			s.clientsLock.Lock()
			delete(s.clients, ip)
			s.clientsLock.Unlock()

			if s.DisconnectedCallback != nil {
				s.DisconnectedCallback(ip)
			}
			return
		default:
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				log.Errorf("Client %s has disconnected", clientAddr)
				s.clientsLock.Lock()
				delete(s.clients, ip)
				s.clientsLock.Unlock()
				return
			}

			log.Printf("Message from client %s: %+v", ip, buffer[:n])

			if s.MessageCallback != nil {
				s.MessageCallback(ip, buffer[:n])
			}
		}
	}
}

func (s *TCPServer) SendMessage(ip string, data []byte) error {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	conn, found := s.clients[ip]
	if !found {
		return fmt.Errorf("client %s not found", ip)
	}

	_, err := conn.Write(data)
	if err != nil {
		return err
	}

	log.Printf("Send message to client %s: %+v", ip, data)
	return nil
}
