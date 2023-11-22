package common_gateway


import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"strings"
	log "github.com/sirupsen/logrus"

	"github.com/mzpbvsig/common-devices-gateway/bean"
)

type DeviceGatewaySearchCallback   func(string, string, int)
type DeviceGatewaySearch2Callback  func(string, string, []string)
type DeviceGatewayTestCallback     func(eventId string, eventMethod string, data string)
type DeviceGatewayRefreshCallback  func()
type DeviceGatewayUnsearchCallback func(gatewayId string)

type RestServer struct {
	router            *mux.Router
	mysqlManager      *MysqlManager
	SearchCallback    DeviceGatewaySearchCallback
	Search2Callback   DeviceGatewaySearch2Callback
	UnsearchCallback  DeviceGatewayUnsearchCallback
	TestCallback      DeviceGatewayTestCallback
	RefreshCallback   DeviceGatewayRefreshCallback
}


func NewRestServer(mysqlManager *MysqlManager) *RestServer {
	s := &RestServer{
		router: mux.NewRouter(),
	}
	s.routes()
	return s
}

func (s *RestServer) routes() {
	s.router.HandleFunc("/test", s.handleTest()).Methods("POST")
	s.router.HandleFunc("/refresh", s.handleRefresh()).Methods("GET")
	s.router.HandleFunc("/search", s.handleSearch()).Methods("GET")
	s.router.HandleFunc("/search2", s.handleSearch2()).Methods("POST")
	s.router.HandleFunc("/unsearch", s.handleUnsearch()).Methods("GET")
	s.router.HandleFunc("/healthz", s.handleHealth()).Methods("GET")
}

// sendJSONResponse is a generic function to send JSON responses.
func sendJSONResponse[T any](w http.ResponseWriter, response bean.ResponseData[T]) {
    // Marshal the response into JSON
    jsonData, err := json.Marshal(response)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error marshalling response: %v", err), http.StatusInternalServerError)
        return
    }

    // Set the header and write the response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(response.Code)
    _, err = w.Write(jsonData)
    if err != nil {
        log.Errorf("Error writing response: %+v", err)
    }
}


func (s *RestServer) handleRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := bean.ResponseData[string]{
			Message: "Refresh request completed",
			Code:    http.StatusOK,
			Data:  "",
		}
		sendJSONResponse(w, response)

		if s.RefreshCallback != nil {
			s.RefreshCallback()
		}
	}
}

func (s *RestServer) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		
		entityId := r.URL.Query().Get("entity_id")
		eventMethod := r.URL.Query().Get("event_method")

		data := ""
		if err := r.ParseForm(); err == nil {
			data = r.FormValue("data")
		}

		if entityId != ""  {
			if s.TestCallback != nil {
				s.TestCallback(entityId, eventMethod, data)
			}
		} else {
			http.Error(w, fmt.Sprintf("%v", "entityId is empty"), http.StatusInternalServerError)
			return
		}
		
		// Create the response struct
		response := bean.ResponseData[string]{
			Message: "Test request completed",
			Code:    http.StatusOK,
			Data:  "",
		}

		sendJSONResponse(w, response)
	}
}

func (s *RestServer) handleSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var deviceGateways []*bean.DeviceGateway
		var err error

		// Parse the query string for gateway_id
		gatewayId := r.URL.Query().Get("gateway_id")
		classId := r.URL.Query().Get("class_id")
		maxSN := r.URL.Query().Get("max_sn")


		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		num, err := strconv.Atoi(maxSN)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return   
		}

		// If a callback is set, pass the device gateways to it
		if s.SearchCallback != nil {
			s.SearchCallback(gatewayId, classId, num)
		}

		// Create the response struct
		response := bean.ResponseData[[]*bean.DeviceGateway]{
			Message: "Search completed", 
			Code:    http.StatusOK,
			Data:    deviceGateways,
		}

		sendJSONResponse(w, response)
	}
}

func (s *RestServer) handleSearch2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var deviceGateways []*bean.DeviceGateway
		var err error

		// Parse the query string for gateway_id
		gatewayId := r.URL.Query().Get("gateway_id")
		classId := r.URL.Query().Get("class_id")

		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		sns := r.FormValue("sns")

		// If a callback is set, pass the device gateways to it
		if s.Search2Callback != nil {	
			s.Search2Callback(gatewayId, classId, strings.Split(sns, ","))
		}

		// Create the response struct
		response := bean.ResponseData[[]*bean.DeviceGateway]{
			Message: "Search completed", 
			Code:    http.StatusOK,
			Data:    deviceGateways,
		}

		sendJSONResponse(w, response)
	}
}

func (s *RestServer) handleUnsearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gatewayId := r.URL.Query().Get("gateway_id")

		// If a callback is set, pass the device gateways to it
		if s.UnsearchCallback != nil {
			// the fuck code will modify on mul protocol
			s.UnsearchCallback(gatewayId)
		}

		// Create the response struct
		response := bean.ResponseData[string]{
			Message: "Search completed", 
			Code:    http.StatusOK,
			Data:    "",
		}

		sendJSONResponse(w, response)
	}
}

func (s *RestServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service is healthy"))
	}
}

func (s *RestServer) Start(port int) {
	log.Printf("RestServer listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), s.router))
}


