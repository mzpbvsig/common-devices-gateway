package bean


type Event struct {
    Id   string `json:"id"`
    Data string `json:"data"`
    Name string `json:"name"`
    Method string `json:"method"`
    Code string  `json:"-"`
}

type EntityClass struct {
    Id      string  `json:"id"`
    Name    string  `json:"name"`
    Events  []Event `json:"events"`
    Method  string  `json:"-"`
    Code    string  `json:"-"`
    Data    string  `json:"-"`
}


type Entity struct {
    Id          string       `json:"id"`
    ClassId     string       `json:"class_id"`
    DeviceId    string       `json:"device_id"`
    Data        string       `json:"data"`
    State       string       `json:"state"`
    EntityClass *EntityClass `json:"entity_class"`
}
