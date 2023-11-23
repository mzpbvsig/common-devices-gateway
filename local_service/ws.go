package local_service

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"fmt"

	"github.com/gorilla/websocket"
)

// WebSocketServer 是 WebSocket 服务器对象
type WebSocketServer struct {
	BaseServer
	clients map[string]*websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源的 WebSocket 连接，您可以自定义跨域策略
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
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 获取客户端地址
	clientAddr := conn.RemoteAddr().String()

	// 将连接保存到 clients 映射中
	s.clientsLock.Lock()
	s.clients[clientAddr] = conn
	s.clientsLock.Unlock()

	log.Printf("Client %s has connected", clientAddr)

	// 处理客户端消息
	for {
		select {
		case <-s.shutdownChan:
			log.Printf("Server is shutting down, client %s has disconnected", clientAddr)
			s.clientsLock.Lock()
			delete(s.clients, clientAddr) // 从 clients 映射中移除连接
			s.clientsLock.Unlock()
			return
		default:
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Errorf("Client %s has disconnected", clientAddr)
				s.clientsLock.Lock()
				delete(s.clients, clientAddr) // 从 clients 映射中移除连接
				s.clientsLock.Unlock()
				return
			}

			log.Printf("Message from client %s: %+v", clientAddr, p)

			if s.MessageCallback != nil {
				s.MessageCallback(clientAddr, p)
			}
		}
	}
}

func (s *WebSocketServer) SendMessage(clientAddr string, message []byte) error {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	conn, found := s.clients[clientAddr]
	if !found {
		return fmt.Errorf("client %s not found", clientAddr)
	}

	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return err
	}

	log.Printf("Send message to client %s: %+v", clientAddr, message)
	return nil
}
