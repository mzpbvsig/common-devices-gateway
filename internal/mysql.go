package internal

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mzpbvsig/common-devices-gateway/bean"
	log "github.com/sirupsen/logrus"
)

type MysqlManager struct {
	db *sql.DB
}

func NewMysqlManager(opt bean.MysqlOption) *MysqlManager {
	db := connect2mysql(opt)

	if db == nil {
		panic("Error to connect to MySQL exit.")
	}

	return &MysqlManager{
		db: db,
	}
}

func connect2mysql(opt bean.MysqlOption) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		opt.UserName,
		opt.AuthPass,
		opt.Ip,
		opt.Port,
		opt.DB,
	)

	var db *sql.DB
	var err error
	maxAttempts := 3
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				return db
			}
		}
		log.Errorf("Attempt %d: Failed to connect to MySQL: %s", attempts, err)
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (manager *MysqlManager) Close() error {
	return manager.db.Close()
}

func (manager *MysqlManager) queryDeviceGateways(optionalGatewayID ...string) ([]*bean.DeviceGateway, error) {
	var rows *sql.Rows
	var err error

	if len(optionalGatewayID) > 0 && optionalGatewayID[0] != "" {
		query := "SELECT id, ip, type, protocol FROM s_gateway WHERE id = ?"
		rows, err = manager.db.Query(query, optionalGatewayID[0])
	} else {
		query := "SELECT id, ip, type, protocol FROM s_gateway"
		rows, err = manager.db.Query(query)
	}

	if err != nil {
		log.Errorf("Failed to retrieve gateways from database: %s", err)
		return nil, err
	}
	defer rows.Close()

	var gateways []*bean.DeviceGateway
	for rows.Next() {
		var gateway bean.DeviceGateway
		err := rows.Scan(&gateway.Id, &gateway.Ip, &gateway.Type, &gateway.Protocol)
		if err != nil {
			log.Errorf("Failed to scan gateway record: %s", err)
			return nil, err
		}
		gateways = append(gateways, &gateway)
	}

	if err := rows.Err(); err != nil {
		log.Errorf("Error while iterating over rows: %s", err)
		return nil, err
	}

	for _, gateway := range gateways {
		devices, err := manager.GetDevices(gateway.Id)
		if err != nil {
			return nil, err
		}
		gateway.Devices = devices
	}

	return gateways, nil
}

func (manager *MysqlManager) GetDeviceGateways() ([]*bean.DeviceGateway, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	return manager.queryDeviceGateways()
}

func (manager *MysqlManager) GetDeviceGatewayByID(id string) (*bean.DeviceGateway, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	gateways, err := manager.queryDeviceGateways(id)
	if err != nil {
		return nil, err
	}
	if len(gateways) > 0 {
		return gateways[0], nil
	}

	return nil, fmt.Errorf("device gateway with ID %s not found", id)
}

func (manager *MysqlManager) AddOrUpdateDevice(device *bean.Device) error {
	if manager.db == nil {
		return fmt.Errorf("database connection is not established")
	}

	stmt, err := manager.db.Prepare(`
		INSERT INTO s_gateway_devices (id, gateway_id, class_id, sn) 
		VALUES (?, ?, ?, ?) 
		ON DUPLICATE KEY UPDATE 
			gateway_id = VALUES(gateway_id), 
			class_id = VALUES(class_id), 
			sn = VALUES(sn)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(device.Id, device.GatewayId, device.ClassId, device.SN)
	if err != nil {
		return err
	}

	return nil
}

func (manager *MysqlManager) UpdateEntity(entity *bean.Entity) error {
	if manager.db == nil {
		return fmt.Errorf("database connection is not established")
	}

	now := time.Now().Unix()

	stmt, err := manager.db.Prepare(`
		UPDATE s_device_entity set state=?, upd_time=? where id=?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(entity.State, now, entity.Id)
	if err != nil {
		return err
	}

	return nil
}

func (manager *MysqlManager) AddDevice(device *bean.Device) error {
	if manager.db == nil {
		return fmt.Errorf("database connection is not established")
	}

	var exists bool
	err := manager.db.QueryRow("SELECT EXISTS(SELECT 1 FROM s_gateway_devices WHERE gateway_id = ? AND sn = ?)", device.GatewayId, device.SN).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err := manager.db.Exec("INSERT INTO s_gateway_devices (gateway_id, class_id, sn) VALUES (?, ?, ?)", device.GatewayId, device.ClassId, device.SN)
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *MysqlManager) queryDevices(query string, args ...interface{}) ([]*bean.Device, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	rows, err := manager.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*bean.Device

	for rows.Next() {
		device := &bean.Device{}
		deviceClass := bean.DeviceClass{}
		err := rows.Scan(&device.Id, &device.GatewayId, &device.ClassId, &device.SN, &deviceClass.Id, &deviceClass.Name, &deviceClass.Protocol, &deviceClass.Model, &deviceClass.Type, &deviceClass.Brand)
		if err != nil {
			return nil, err
		}
		device.DeviceClass = &deviceClass
		device.Entities, err = manager.GetEntities(device.Id)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (manager *MysqlManager) GetAllDevices() ([]*bean.Device, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := `SELECT d.id, d.gateway_id, d.class_id, d.sn, dc.id, dc.name, p.name as protocol, dc.model,dc.type, dc.brand
			  FROM s_gateway_devices d
			  JOIN s_device_class dc ON d.class_id = dc.id
			  JOIN s_device_protocol p ON dc.protocol_id = p.id`

	return manager.queryDevices(query)
}

func (manager *MysqlManager) GetDevices(gatewayId string) ([]*bean.Device, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := `SELECT d.id, d.gateway_id, d.class_id, d.sn, dc.id, dc.name, p.name as protocol, dc.model,dc.type, dc.brand
			  FROM s_gateway_devices d
			  JOIN s_device_class dc ON d.class_id = dc.id
			  JOIN s_device_protocol p ON dc.protocol_id = p.id
			  WHERE d.gateway_id=?`
	return manager.queryDevices(query, gatewayId)
}

func (manager *MysqlManager) GetEntities(deviceId string) ([]*bean.Entity, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := `
        SELECT e.id, e.class_id, e.device_id, e.state, e.data, ec.id, ec.name, ec.method, ec.data , ec.code
        FROM s_device_entity e
        JOIN s_device_entity_class ec ON e.class_id = ec.id
        WHERE e.device_id=?
    `
	rows, err := manager.db.Query(query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*bean.Entity
	for rows.Next() {
		var entity bean.Entity
		entity.EntityClass = &bean.EntityClass{}
		var tempCode sql.NullString
		err := rows.Scan(
			&entity.Id, &entity.ClassId, &entity.DeviceId, &entity.State, &entity.Data,
			&entity.EntityClass.Id, &entity.EntityClass.Name, &entity.EntityClass.Method, &entity.EntityClass.Data, &tempCode,
		)
		if err != nil {
			return nil, err
		}
		if tempCode.Valid {
			entity.EntityClass.Code = tempCode.String
		} else {
			entity.EntityClass.Code = ""
		}
		entities = append(entities, &entity)

		entity.EntityClass.Events, err = manager.GetEventsByEntityClassId(entity.EntityClass.Id)

		if err != nil {
			return nil, err
		}
	}

	return entities, nil
}

func (manager *MysqlManager) LoadDeviceGateways() ([]*bean.DeviceGateway, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	gateways, err := manager.GetDeviceGateways()
	if err != nil {
		return nil, err
	}

	return gateways, nil
}

func (manager *MysqlManager) LoadDeviceClasses() ([]bean.DeviceClass, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	rows, err := manager.db.Query(`SELECT dc.id, dc.name, p.name as protocol, dc.model, dc.brand, dc.type FROM s_device_class dc
		JOIN s_device_protocol p ON dc.protocol_id = p.id
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deviceClasses []bean.DeviceClass
	for rows.Next() {
		dc := bean.DeviceClass{}
		err := rows.Scan(&dc.Id, &dc.Name, &dc.Protocol, &dc.Model, &dc.Brand, &dc.Type)
		if err != nil {
			return nil, err
		}

		dc.EntityClasses, err = manager.GetEntitiyClassesByDeviceClassId(dc.Id)
		if err != nil {
			return nil, err
		}

		deviceClasses = append(deviceClasses, dc)
	}

	return deviceClasses, nil
}

func (manager *MysqlManager) GetEntitiyClassesByDeviceClassId(deviceClassId string) ([]bean.EntityClass, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	rows, err := manager.db.Query("SELECT id, name, method, data, code FROM s_device_entity_class WHERE device_cid=?", deviceClassId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entitiyClasses []bean.EntityClass
	for rows.Next() {
		entityClass := bean.EntityClass{}
		var tempCode sql.NullString
		err := rows.Scan(&entityClass.Id, &entityClass.Name, &entityClass.Method, &entityClass.Data, &tempCode)
		if err != nil {
			return nil, err
		}

		if tempCode.Valid {
			entityClass.Code = tempCode.String
		} else {
			entityClass.Code = ""
		}

		entityClass.Events, err = manager.GetEventsByEntityClassId(entityClass.Id)

		if err != nil {
			return nil, err
		}

		entitiyClasses = append(entitiyClasses, entityClass)
	}
	return entitiyClasses, nil
}

func (manager *MysqlManager) GetEventsByEntityClassId(entityClassId string) ([]bean.Event, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	rows, err := manager.db.Query("SELECT id, data, name, method, code FROM s_device_entity_events WHERE entity_class_id=?", entityClassId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []bean.Event

	for rows.Next() {
		event := bean.Event{}

		var tempCode sql.NullString

		err := rows.Scan(&event.Id, &event.Data, &event.Name, &event.Method, &tempCode)
		if err != nil {
			return nil, err
		}

		if tempCode.Valid {
			event.Code = tempCode.String
		} else {
			event.Code = ""
		}

		events = append(events, event)

	}

	return events, nil
}

func (manager *MysqlManager) UpdateSearchDone(gatewayID string) error {
	if manager.db == nil {
		return fmt.Errorf("database connection is not established")
	}

	now := time.Now().Unix()

	stmt := "UPDATE s_search_log SET search_time = ?, process=0, is_done = 1 WHERE gateway_id = ?"

	_, err := manager.db.Exec(stmt, now, gatewayID)

	return err
}

func (manager *MysqlManager) UpdateSearchProcess(gatewayId string, process string) error {
	if manager.db == nil {
		return fmt.Errorf("database connection is not established")
	}

	stmt := "UPDATE s_search_log SET process = ?, is_done = 0 WHERE gateway_id = ?"

	_, err := manager.db.Exec(stmt, process, gatewayId)

	return err
}

func (manager *MysqlManager) AddLog(timestamp time.Time, message string, level string) error {
	query := "INSERT INTO s_run_logs (timestamp, message, level) VALUES (?, ?, ?)"

	_, err := manager.db.Exec(query, timestamp, message, level)
	if err != nil {
		return err
	}

	return nil
}

func (manager *MysqlManager) GetAllProtocols() ([]*bean.Protocol, error) {
	if manager.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	rows, err := manager.db.Query("select id, name, request_code, response_code from s_device_protocol")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var protocols []*bean.Protocol

	for rows.Next() {
		protocol := &bean.Protocol{}
		err := rows.Scan(&protocol.Id, &protocol.Name, &protocol.RequestCode, &protocol.ResponseCode)
		if err != nil {
			return nil, err
		}
		protocols = append(protocols, protocol)
	}

	return protocols, nil
}
