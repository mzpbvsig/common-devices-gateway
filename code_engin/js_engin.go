package code_engin

import (
	"fmt"

	"github.com/dop251/goja"
)

type JSEngine struct {
	vm *goja.Runtime
}

func NewJSEngine() *JSEngine {
	return &JSEngine{
		vm: goja.New(),
	}
}

func (engine *JSEngine) RunJs(jsCode string, data []byte) (string, error) {
	var jsData []interface{}
	for _, b := range data {
		jsData = append(jsData, byte(b))
	}
	engine.vm.Set("data", jsData)

	code := fmt.Sprintf(`
        function getinfo(data){
            %s
        }
        var result = getinfo(data);
    `, jsCode)

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
