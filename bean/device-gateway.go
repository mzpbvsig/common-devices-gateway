package bean

type DeviceGateway struct {
	Id       string    `json:"id"`
	SN       string    `json:"SN"`
	Ip       string    `json:"ip"`
	Type     string    `json:"type"`
	Devices  []*Device `json:"devices"`
	Protocol string    `json:"protocol"`
	IsOnline bool      `json:"is_online"`
}
