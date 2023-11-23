package code_engin

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/mzpbvsig/common-devices-gateway/bean"
)

type JSEngine struct {
	vm *goja.Runtime
}

func NewJSEngine() *JSEngine {
	return &JSEngine{
		vm: goja.New(),
	}
}

func (engine *JSEngine) Request(jsCode string, device *bean.Device, entity *bean.Entity) ([]byte, error) {
	engine.vm.Set("device", device)
	engine.vm.Set("entity", entity)

	code := fmt.Sprintf(`
        function request(){
            %s
        }
        var result = request();
    `, jsCode)

	_, err := engine.vm.RunString(code)
	if err != nil {
		return nil, err
	}
	value := engine.vm.Get("result")
	if arr, ok := value.Export().([]interface{}); ok {
		var byteArr []byte
		for _, v := range arr {
			if num, ok := v.(int64); ok {
				byteArr = append(byteArr, byte(num))
			}
		}
		return byteArr, nil
	}
	return nil, fmt.Errorf("cannot convert result to array")

}

func (engine *JSEngine) Response(jsCode string, device *bean.Device, entity *bean.Entity, response []byte) (string, error) {
	engine.vm.Set("device", device)
	engine.vm.Set("entity", entity)
	var responseData []interface{}
	for _, b := range response {
		responseData = append(responseData, byte(b))
	}
	engine.vm.Set("data", responseData)

	code := fmt.Sprintf(`
        function response(){
            %s  
            function getinfo(){
                %s
            } 
            return getinfo(params);
        }
        var result = response();
    `, jsCode, entity.EntityClass.Code)

	_, err := engine.vm.RunString(code)
	if err != nil {
		return "", err
	}

	value := engine.vm.Get("result")
	if result, ok := value.Export().(string); ok {
		return result, nil
	}
	return "", fmt.Errorf("cannot convert result to string")
}
