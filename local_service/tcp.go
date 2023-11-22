package local_service

import (
	"net"
	"sync"
	"strings"
	"fmt"
	"errors"
	log "github.com/sirupsen/logrus"
)

// TCPServer is a TCP server object
type TCPServer struct {
	listener        net.Listener
	clients         map[string]net.Conn
	clientsLock     sync.Mutex
	shutdownChan    chan struct{}
	MessageCallback func(clientAddr string, data []byte)
	ConnectedCallback func(clientAddr string)
	DisconnectedCallback func(clientAddr string)
}

// NewTCPServer creates a new TCPServer object
func NewTCPServer(messageCallback func(clientAddr string, data []byte), connectedCallback func(clientAddr string),	disconnectedCallback func(clientAddr string)) *TCPServer {
	return &TCPServer{
		clients:         make(map[string]net.Conn),
		shutdownChan:    make(chan struct{}),
		MessageCallback: messageCallback,
		ConnectedCallback: connectedCallback,
		DisconnectedCallback: disconnectedCallback,
	}
}

// Start starts the TCP server
func (s *TCPServer) Start(port int) {
	// Start the TCP server
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Println("Failed to start the server:", err)
		return
	}
	defer s.listener.Close()

	log.Println("Server has started, waiting for client connections...")

	for {
		// Wait for client connections
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Client connection error:", err)
			continue
		}

		// Handle client connection
		go s.handleClient(conn)
	}
}

// Stop stops the TCP server
func (s *TCPServer) Stop() {
	close(s.shutdownChan)
	s.listener.Close()
}

func (s *TCPServer) handleClient(conn net.Conn) {
	// Get the client address
	clientAddr := conn.RemoteAddr().String()

	// Save the client connection to the clients map
	s.clientsLock.Lock()
	s.clients[clientAddr] = conn
	s.clientsLock.Unlock()

	log.Printf("Client %s has connected", clientAddr)
	if s.ConnectedCallback != nil {
		s.ConnectedCallback(clientAddr)
	}
	// Handle client messages
	for {
		select {
		case <-s.shutdownChan:
			log.Printf("Server is shutting down, client %s has disconnected", clientAddr)
			s.clientsLock.Lock()
			delete(s.clients, clientAddr) // Remove the connection from the clients map
			s.clientsLock.Unlock()

			if s.DisconnectedCallback != nil {
				s.DisconnectedCallback(clientAddr)
			}
			return
		default:
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				log.Errorf("Client %s has disconnected", clientAddr)
				s.clientsLock.Lock()
				delete(s.clients, clientAddr) // Remove the connection from the clients map
				s.clientsLock.Unlock()
				return
			}

			log.Printf("Message from client %s: %+v", clientAddr, buffer[:n])

			if s.MessageCallback != nil {
				s.MessageCallback(clientAddr, buffer[:n])
			}
		}
	}
}

func (s *TCPServer) SendMessage(senderAddr string, data []byte) error {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	var errs []string
	for clientAddr, conn := range s.clients {
		if strings.Index(clientAddr, senderAddr) != -1 {
			_, err := conn.Write(data)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to send message to client %s: %+v", clientAddr, err)
				log.Error(errMsg)
				errs = append(errs, errMsg)
			} else {
				log.Printf("Sent message to client %s: %+v", clientAddr, data)
			}
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

