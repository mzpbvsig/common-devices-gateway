# 通用网关程序设计文档

## 介绍

这份文档描述了一个通用网关程序的设计和组织结构。该程序的主要目标是允许与多种设备通信，收集数据，并将数据发送到云端服务。程序的核心部分由 Golang 编写，采用了模块化的设计，以便易于扩展和维护。

## 目录结构

通用网关程序的源代码按以下目录结构组织：

- **main.go**: 主程序入口点，负责启动各个服务模块并协调它们之间的工作。

- **bean**: 存放实体对象的包，包括设备网关、设备、实体、事件以及配置文件等。这些实体对象用于描述设备和配置。

- **cloud_service**: 云端服务模块，负责与云端通信，包括消息传递、Pulsar 集成、REST 协议等。这个模块处理与云端的数据交互。

- **local_service**: 本地服务模块，提供给设备网关的本地服务，可以是 TCP、WebSocket、MQTT 等通信协议。这个模块用于与设备通信。

- **business**: 云端和本地服务的业务封装模块

- **data_struct**: 存放自定义数据结构的包，包括队列等自定义数据结构，用于数据的存储和传递。

- **protocol**: 设备传输数据的协议模块，目前包括 Modbus，可以扩展以支持新的协议。

- **devices**: 设备数据处理模块，包括设备具体的数据业务逻辑，以及设备数据解析器。这个模块用于处理从设备接收到的数据，并将其转换为可用于云端的数据格式。此外，`devices` 目录下包含子文件夹，每个子文件夹可以对一类品牌或相同通信协议的设备进行复用，例如 `vms\modbus` 子文件夹。

- **utils**: 常用的工具函数和辅助函数存放在这个包中，用于程序的辅助功能。

- **test**: 测试用例和测试数据存放在这个目录下，用于测试程序的各个模块。

- **config**: 存放设备网关的配置文件，每个设备网关可以有一个独立的配置文件，例如 `gateway_1.yaml`。

## 设计思路

以下是通用网关程序的设计思路：

1. **模块化设计**: 整个程序采用模块化的设计，每个模块负责特定的功能。这使得程序易于维护和扩展。

2. **设备描述**: 通过 `bean` 包中的实体对象，对设备进行详细描述，包括设备类型、配置、实体、事件等信息。这些描述可用于动态配置和管理设备。

3. **云端服务**: `cloud_service` 模块负责与云端通信，包括消息传递、Pulsar 集成、REST API 等。这使得程序能够将采集的数据发送到云端进行进一步处理。

4. **本地服务**: `local_service` 模块提供本地通信服务，可以支持多种通信协议，包括 TCP、WebSocket、MQTT 等。这个模块负责与设备通信，接收数据并传递给数据处理模块。

5. **数据处理**: `devices` 模块负责具体设备的数据处理逻辑，包括数据解析和业务逻辑处理。这个模块允许添加新的设备，并为每个设备定义特定的数据处理方法。

6. **数据协议**: `protocol` 设备传输数据的协议模块，目前包括 Modbus，可以扩展以支持新的协议。

7. **自定义数据结构**: `data_struct` 包中定义了一些自定义的数据结构，例如队列，用于数据的存储和传递。

8. **配置管理**: 配置文件存放在 `config` 目录下，每个设备网关可以有独立的配置文件。程序会根据配置文件动态配置设备和服务。

## 主程序逻辑

### 基础设置

在主程序 `business.go` 中，首先进行了一些基础设置和初始化工作：

- 加载基础配置文件：通过 `loadConfig()` 函数加载配置文件，包括设备网关、Pulsar 服务等的配置信息。

- 初始化队列和设备处理器：创建一个数据队列和设备处理器，用于处理设备数据的存储和处理。

- 同步的 `nextSendDataChan` 和退出 `stopChan`：这两个通道用于协调数据发送和程序退出的操作。

### 云端和本地服务初始化

接下来，程序初始化了云端服务和本地服务：

- 初始化云端服务：通过 `NewCloudServer(config, dp, queue)` 函数初始化所有支持的服务，用于连接云端服务。

- 初始化本地服务器：通过 `NewLocalServer(handleTCPData)` 函数初始化本地所有支持的服务，用于接收来自设备的数据。

### 定时数据构造和发送

程序启动了两个后台协程来处理定时数据构造和发送：

- 定时数据构造：通过 `makeDeviceDataLoop()` 函数，程序定期构造设备数据并将其添加到队列中。

- 发送定时构造：通过 `sendDeviceDataLoop()` 函数，程序定期从队列中取出数据并发送给设备。

### 加载设备网关

最后，程序加载设备网关及其下设备：

- 加载设备网关配置文件：通过 `loadDeviceGateways()` 函数加载设备网关的配置文件，并将其注册到 Pulsar 云端服务。

### 阻塞主线程

最后，主程序使用 `select{}` 语句来阻塞主线程，以保持程序的运行。程序将一直运行，直到接收到退出信号。

这个程序的主要任务是接收、处理和发送设备数据，它以高度模块化和可扩展的方式组织代码，使得可以轻松添加新的设备和功能。


## 如何扩展

通用网关程序的设计允许轻松扩展以下方面：

1. **添加新设备**: 要添加新设备，只需在 `devices` 模块中定义新的设备处理逻辑，并在配置文件中描述设备的配置。程序将根据配置自动识别和处理新设备。

2. **添加新协议**: 若要添加新的通信协议，只需在 `protocol` 模块中实现新协议的解析和数据构建方法。程序将能够处理多种协议。

3. **扩展云端服务**: 可以通过扩展 `cloud_service` 模块来支持更多云端服务，例如支持其他消息系统或云端 API。

4. **改进本地服务**: 通过扩展 `local_service` 模块，可以添加新的本地通信协议或改进现有协议的性能和功能。

5. **新增实体和事件**: 随着设备的不断发展，可以添加新的实体和事件，以满足新的需求。

通过以上扩展方式，通用网关程序可以适应不断变化的设备和通信需求，使其更具灵活性和可维护性。

## 结论

通用网关程序是一个灵活且可扩展的解决方案，可用于连接各种设备并将其数据发送到云端服务。其模块化的设计允许轻松添加新设备、协议和功能，使其成为一个通用的物联网网关解决方案。