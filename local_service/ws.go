package local_service

import (
	"net/http"

	"github.com/mzpbvsig/common-devices-gateway/utils"
	log "github.com/sirupsen/logrus"

	"fmt"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	BaseServer
	clients map[string]*websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketServer(messageCallback func(clientAddr string, data []byte), connectedCallback func(clientAddr string), disconnectedCallback func(clientAddr string)) *WebSocketServer {
	return &WebSocketServer{
		clients: make(map[string]*websocket.Conn),
		BaseServer: BaseServer{
			shutdownChan:         make(chan struct{}),
			MessageCallback:      messageCallback,
			ConnectedCallback:    connectedCallback,
			DisconnectedCallback: disconnectedCallback,
		},
	}
}

func (s *WebSocketServer) Start(port int, path string) {
	http.HandleFunc(path, s.handleWebSocket)
	go s.listenAndServe(port)
}

func (s *WebSocketServer) Stop() {
	close(s.shutdownChan)
}

func (s *WebSocketServer) listenAndServe(port int) {
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade error:", err)
		return
	}
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	log.Printf("Client %s has connected", clientAddr)

	ip := utils.ExtractIP(clientAddr)

	s.clientsLock.Lock()
	s.clients[ip] = conn
	s.clientsLock.Unlock()

	for {
		select {
		case <-s.shutdownChan:
			log.Printf("Server is shutting down, client %s has disconnected", clientAddr)
			s.clientsLock.Lock()
			delete(s.clients, ip)
			s.clientsLock.Unlock()
			return
		default:
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Errorf("Client %s has disconnected", clientAddr)
				s.clientsLock.Lock()
				delete(s.clients, ip)
				s.clientsLock.Unlock()
				return
			}

			log.Printf("Message from client %s: %+v", clientAddr, p)

			if s.MessageCallback != nil {
				s.MessageCallback(ip, p)
			}
		}
	}
}

func (s *WebSocketServer) SendMessage(ip string, message []byte) error {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	conn, found := s.clients[ip]
	if !found {
		return fmt.Errorf("client %s not found", ip)
	}

	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return err
	}

	log.Printf("Send message to client %s: %+v", ip, message)
	return nil
}
