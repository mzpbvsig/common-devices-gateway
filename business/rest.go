package business

import (
	"fmt"
	"strings"

	"github.com/mzpbvsig/common-devices-gateway/bean"
	"github.com/mzpbvsig/common-devices-gateway/internal"
	log "github.com/sirupsen/logrus"
)

type SearchCallback func(*bean.Device, *bean.Entity)
type TestClassback func(*bean.Entity)

type RestManager struct {
	SearchCallback SearchCallback
	TestClassback  TestClassback
	deviceMap      map[string]map[string]bool
	searchTotals   map[string]int
	searchNums     map[string]int
}

func NewRestManager(searchCB SearchCallback, testCB TestClassback) *RestManager {
	return &RestManager{
		SearchCallback: searchCB,
		TestClassback:  testCB,
		deviceMap:      make(map[string]map[string]bool),
		searchTotals:   make(map[string]int),
		searchNums:     make(map[string]int),
	}
}

func (restManager *RestManager) Start() {
	restServer := internal.NewRestServer(mysqlManager)
	restServer.SearchCallback = restManager.search
	restServer.Search2Callback = restManager.search2
	restServer.TestCallback = restManager.test
	restServer.RefreshCallback = restManager.refresh
	restServer.UnsearchCallback = restManager.unsearch
	go restServer.Start(config.RestPort)
}

func (restManager *RestManager) Increasing(gatewayId string) {
	restManager.searchNums[gatewayId]++
}

// search searches for devices and processes their data
func (restManager *RestManager) performSearch(rgatewayId string, classId string, snProvider func(int) string, snCount int) {
	for _, deviceGateway := range config.DeviceGateways {
		gatewayId := deviceGateway.Id

		if rgatewayId == "" || rgatewayId == "0" {

		} else if rgatewayId != gatewayId {
			continue
		}

		if !deviceGateway.IsOnline {
			log.Warnf("Search deviceGateway %s is offline", deviceGateway.Ip)
			continue
		}
		if restManager.deviceMap[deviceGateway.Id] == nil {
			restManager.deviceMap[deviceGateway.Id] = make(map[string]bool)
		} else {
			log.Warnf("Searching deviceGateway id is %s", gatewayId)
			continue
		}
		restManager.searchNums[gatewayId] = 0
		restManager.searchTotals[gatewayId] = 0
		for _, deviceClass := range config.DeviceClasses {
			if classId == "0" || classId == "" {

			} else if classId != deviceClass.Id {
				continue
			}

			if len(deviceClass.EntityClasses) > 0 {
				restManager.searchTotals[gatewayId] += snCount
			}
		}

		log.Printf("Search total count is %d", restManager.searchTotals[gatewayId])

		for i := snCount; i >= 1; i-- {
			SN := snProvider(i)
			for _, deviceClass := range config.DeviceClasses {
				currentDeviceClass := deviceClass

				if classId == "0" || classId == "" {

				} else if classId != deviceClass.Id {
					continue
				}

				device := &bean.Device{}
				device.Id = "0"
				device.GatewayId = deviceGateway.Id
				device.DeviceClass = &currentDeviceClass
				device.ClassId = deviceClass.Id
				device.SN = SN

				for _, entityClass := range device.DeviceClass.EntityClasses {

					if !strings.HasPrefix(strings.ToLower(entityClass.Method), "get") {
						continue
					}

					entity := &bean.Entity{}
					entity.ClassId = entityClass.Id
					entity.EntityClass = &entityClass
					entity.DeviceId = device.Id
					entity.Id = "0"
					data, err := dp.ProcessMakeDeviceData(device, entity)

					if err != nil {
						log.Errorf("ProcessMakeDeviceData err is %s", err)
						continue
					}

					sendData := &DeviceData{
						DeviceGateway: deviceGateway,
						Data:          data,
						Entity:        entity,
						Device:        device,
						Type:          Search,
					}
					dataManager.Unshift(deviceGateway.Id, sendData)

					break
				}
			}
		}
		if restManager.searchTotals[gatewayId] == 0 {
			restManager.reset(gatewayId)
		}

	}
}

func (restManager *RestManager) search(gatewayId string, classId string, maxSn int) {
	snProvider := func(i int) string {
		return fmt.Sprintf("%d", i)
	}
	restManager.performSearch(gatewayId, classId, snProvider, maxSn)
}

func (restManager *RestManager) search2(gatewayId string, classId string, sns []string) {
	snProvider := func(i int) string {
		return sns[i-1]
	}
	restManager.performSearch(gatewayId, classId, snProvider, len(sns))
}

func (restManager *RestManager) reset(gatewayId string) {
	restManager.searchNums[gatewayId] = 0
	restManager.searchTotals[gatewayId] = 0
	restManager.deviceMap[gatewayId] = nil
}

func (restManager *RestManager) searched(device *bean.Device, entity *bean.Entity) {
	if restManager.SearchCallback != nil {
		restManager.SearchCallback(device, entity)
	}
}

func (restManager *RestManager) isSearchDone(deviceData *DeviceData) bool {
	if deviceData == nil {
		return false
	}

	if deviceData.Type != Search {
		return false
	}

	if deviceData.Device != nil {
		gatewayId := deviceData.Device.GatewayId
		restManager.Increasing(deviceData.Device.GatewayId)
		log.Printf("Search num is %d total is %d", restManager.searchNums[gatewayId], restManager.searchTotals[gatewayId])
		return restManager.checkSearchCompletion(gatewayId)
	}

	return false
}

func (restManager *RestManager) checkSearchCompletion(gatewayId string) bool {
	currentSearchNum := restManager.searchNums[gatewayId]
	totalSearches := restManager.searchTotals[gatewayId]
	process := (float64(currentSearchNum) / float64(totalSearches)) * 100.0
	if process >= 100.0 {
		return true
	} else {
		log.Printf("Search process is %.2f", process)
		restManager.UpdateSearchProcess(gatewayId, fmt.Sprintf("%.2f", process))
		return false
	}
}

func (restManager *RestManager) UpdateSearchProcess(gatewayId string, process string) {
	mysqlManager.UpdateSearchProcess(gatewayId, process)
}

func (restManager *RestManager) test(entityId string, eventMethod string, data string) {
	err := dispatchEvent(entityId, eventMethod, data, Test)
	if err != nil {
		log.Errorf("Test make data error: %+v ", err)
	}
}

func (restManager *RestManager) refresh() {
	loadDeviceGateways()

	dataManager.BuildQuickDeviceDatas()

	cloudServer.Registers()
	cloudServer.CreateStateProducers()

	log.Printf("Refresh data [loadDeviceGateways BuildQuickDeviceDatas Registers CreateStateProducers]")
}

func (restManager *RestManager) unsearch(gatewayId string) {
	if gatewayId == "0" || gatewayId == "" {
		for _, gateway := range config.DeviceGateways {
			unsearch(gateway.Id)
		}
	} else {
		unsearch(gatewayId)
	}

	log.Printf("Undo search")
}

func unsearch(gatewayId string) {
	dataManager.RemoveAll(gatewayId)
	restManager.reset(gatewayId)
	mysqlManager.UpdateSearchDone(gatewayId)
	loadDevices(gatewayId)
}
