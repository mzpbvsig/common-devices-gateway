package business

import (
	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/devices"
	internal "github.com/mzpbvsig/common-devices-gateway/internal"
	"github.com/mzpbvsig/common-devices-gateway/loghook"

	log "github.com/sirupsen/logrus"
)

// Global configuration and state variables
var (
	config       bean.Config
	mysqlManager *internal.MysqlManager
	restManager  *RestManager
	dataManager  *DataManager
	dp           *devices.DeviceProcessor

	nextSendDataChan chan bool
	stopChan         chan struct{}
	cloudServer      *CloudServer
	localServer      *LocalServer
)

// Start initializes and starts the main business logic
func Start() {

	// Load configuration and initialize device gateways
	loadConfig()

	mysqlManager = internal.NewMysqlManager(config.MysqlOption)

	dbHook := loghook.NewDatabaseHook(mysqlManager)
	log.AddHook(dbHook)

	log.Printf("Loaded configuration: %+v", config)

	loadDeviceGateways()
	loadDeviceClasses()

	// Search device callback
	searchCallback := func(device *bean.Device, entity *bean.Entity) {
		log.Printf("Search Callback: Device %+v  Entity %+v", *device, *entity)
		err := mysqlManager.AddDevice(device)
		if err != nil {
			log.Errorf("AddDevice err: %+v", err)
		}
	}

	testCallback := func(entity *bean.Entity) {
		log.Printf("Test Callback :%+v", *entity)
		err := mysqlManager.UpdateEntity(entity)
		if err != nil {
			log.Errorf("UpdateEntity err: %+v", err)
		}
	}

	restManager = NewRestManager(searchCallback, testCallback)
	restManager.Start()

	// Initialize data structures and channels
	dataManager = NewDataManager()
	dp = devices.NewDeviceProcessor()

	nextSendDataChan = make(chan bool)
	stopChan = make(chan struct{})

	// Start cloud server
	cloudServer = NewCloudServer(config)
	cloudServer.Registers()

	// Start local server
	localServer = NewLocalServer(config, handleData, handleConnected, handleDisconnected)

	sendDeviceDataLoop()

	// Wait for termination signal
	select {
	case <-stopChan:
		log.Println("Program terminated.")
	}
}
