# 通用网关程序设计文档

## 介绍

这份文档描述了一个通用网关程序的设计和结构。该程序的主要目标是允许与多种设备通信，收集数据，并将数据发送到云端服务。程序的核心部分由 Golang 编写，采用了模块化的设计，以便易于扩展和维护。

## 目录结构

通用网关程序的源代码按以下目录结构组织：

- **main.go**: 主程序入口点，负责启动各个服务模块并协调它们之间的工作。

- **bean**: 存放实体对象的包，包括设备网关、设备、实体、事件以及配置文件等。这些实体对象用于描述设备和配置。

- **cloud_service**: 云端服务模块，负责与云端通信，包括消息传递、Pulsar 集成、REST 协议等。这个模块处理与云端的数据交互。

- **local_service**: 本地服务模块，提供给设备网关的本地服务，可以是 TCP、WebSocket 等通信协议。这个模块用于与设备通信。

- **business**: 云端和本地服务的业务封装模块

- **internal**: 加载配置, 提供内部rest服务

- **loghook**:  日志入库的hook模块

- **data_struct**: 存放自定义数据结构的包，包括队列等自定义数据结构，用于数据的存储和传递。

- **code_engin**: 协议js代码解析引擎， 可以后台添加js脚本代码。

- **devices**: 设备数据处理模块，转发到协议js代码解析引擎

- **utils**: 常用的工具函数和辅助函数存放在这个包中，用于程序的辅助功能。

- **test**: 测试用例和测试数据存放在这个目录下，用于测试程序的各个模块。

## 协议js代码解析引擎

### 协议请求代码

协议请求代码的主要任务是构造与设备通信的协议数据。该过程涉及以下几个步骤：

1. **输入参数**：该代码部分接收`device`和`entity`对象以及`requestData`对象。
2. **数据构造**：基于输入参数，JS代码需要构造一个请求数据包, 根据协议规范构造请求。
3. **返回值**：JS代码应返回一个数组，代表构造的协议数据。

示例代码结构：

```javascript
function protocol(){
    let request = {};
    try {
         requestData = JSON.parse(entity.EntityClass.Data);
    } catch(e){
         console.log(e);
    }
    // 构造请求数据的代码
    // 返回数据数组
}
var result = protocol();
```

### 协议响应代码

协议响应代码的主要任务是验证设备响应的数据的有效性和一致性。该过程涉及以下几个步骤：

1. **输入参数**：该代码部分接收`device`和`entity`对象以及`requestData`对象以及data（设备响应的数据）
2. **数据验证**：基于输入参数，JS代码需要验证响应数据的有效性和一致性。
3. **返回值**：JS代码应返回一个字符串，表明验证结果。
```javascript
function protocol(){
    let request = {};
    try {
         request = JSON.parse(entity.EntityClass.Data);
    } catch(e){
         console.log(e);
    }
    // 验证数据的代码
}

var result = protocol();

```

### 协议数据解析代码

数据解析代码的主要任务是根据设备响应的数据进行数据处理，按照协议文档执行数据解析。该过程涉及以下几个步骤：

1. **输入参数**： 该代码部分接收`data`，即设备响应的数据,`params`解析用到的参数
2. **数据处理**：基于输入参数，JS代码需要按照协议文档处理数据。
3. **返回值**：JS代码应返回一个字符串，代表处理后的数据结果。

```javascript
let params= undefined
 if(requestData.params){
    params = JSON.parse(requestData.params)
}
function parseData(){
    // 按照协议文档处理数据的代码
    // 返回处理结果
}
 return parseData(params);
```

遵循这些编写规范，开发者可以有效地创建用于通用网关程序的JS代码，从而确保数据的正确传输和处理。这将极大地提升物联网网关的可靠性和灵活性。

## 结论

通用网关程序是一个灵活且可扩展的解决方案，可用于连接各种设备并将其数据发送到云端服务。其模块化的设计允许轻松添加新设备、协议和功能，使其成为一个通用的物联网网关解决方案。