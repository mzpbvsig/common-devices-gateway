package bean

type DeviceClass struct {
    Id            string        `json:"id"`
    Name          string        `json:"name"`
    Protocol      string        `json:"protocol"`
    Model         string        `json:"model"`
    Brand         string        `json:"brand"`
    Type          string        `json:"type"`
    EntityClasses []EntityClass `json:"-"`
}

type Device struct {
    Id          string         `json:"id"`
    GatewayId   string         `json:"gateway_id"`
    ClassId     string         `json:"class_id"`
    SN          string         `json:"sn"`
    Entities    []*Entity      `json:"entities"`
    DeviceClass *DeviceClass   `json:"device_class"`
}
