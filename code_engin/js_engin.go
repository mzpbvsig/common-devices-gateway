package code_engin

import (
    "fmt"
    "github.com/dop251/goja"
)

// JSEngine 类型代表 JavaScript 代码执行引擎
type JSEngine struct {
    vm *goja.Runtime
}

// NewJSEngine 创建并返回一个新的 JSEngine 实例
func NewJSEngine() *JSEngine {
    return &JSEngine{
        vm: goja.New(),
    }
}

// RunJs 执行 JavaScript 代码并返回结果
func (engine *JSEngine) RunJs(jsCode string , data []byte) (string, error) {
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
    return "", fmt.Errorf("Cannot convert result to string")
}
