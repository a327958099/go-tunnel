# Go Tunnel

[English](README.md)

一个基于Go实现的极其轻量级的内网穿透工具，专为在开发阶段将外部网络请求无缝转发至本地服务而设计，用于接收支付回调、信息通知等场景。

## 特性

- 支持 http, https, ws, wss 协议。
- 使用单个 `tunnel-server-config.yaml` 文件轻松配置。
- 使用用户名/密码验证确保客户端-服务器通信安全。
- 无运行时依赖,程序可直接执行,无需安装任何额外的软件。

## 架构设计

### 核心通信机制

Go Tunnel 采用透明的TCP数据桥接架构，支持任何基于TCP的协议，无需针对特定协议进行处理：

```
外部客户端 ↔ 服务端 ↔ 代理连接 ↔ 客户端 ↔ 本地服务
```

**核心组件：**

1. **控制连接**：处理客户端与服务端之间的认证和管理
2. **代理连接池**：预先建立的连接（默认20个）等待处理传入请求
3. **透明数据桥接**：使用 `io.Copy()` 实现零开销的字节级数据转发

**数据流转过程：**

1. **外部请求**：服务端接收来自外部客户端的连接
2. **连接分配**：服务端从连接池中分配一个可用的代理连接
3. **透明转发**：服务端使用 `io.Copy()` 桥接外部连接 ↔ 代理连接
4. **本地转发**：客户端使用 `io.Copy()` 桥接代理连接 ↔ 本地服务

**技术优势：**

- **协议无关**：支持HTTP/HTTPS/WebSocket及任何基于TCP的协议
- **零开销**：直接字节流转发，无额外处理开销
- **高性能**：利用Go语言高效的I/O复制机制
- **透明传输**：保持原始协议完整性，无需修改

## 支持的平台

- Windows
- macOS
- Linux

## 直接下载使用
下载链接：[https://github.com/a327958099/go-tunnel/releases](https://github.com/a327958099/go-tunnel/releases)

## 从源码构建

1.  克隆仓库：
    ```bash
    git clone https://github.com/a327958099/go-tunnel.git
    cd go-tunnel
    ```

2.  构建服务端和客户端：
    ```bash
    go build -o tunnel-server ./cmd/server
    go build -o tunnel-client ./cmd/client
    ```
    这将在项目根目录中创建 `tunnel-server` 和 `tunnel-client` 可执行文件。

## 使用

### 服务端

1.  运行服务端：

    ```bash
    ./tunnel-server
    ```

2.  在服务端可执行文件所在的同一目录中会自动创建一个 `tunnel-server-config.yaml` 文件，内容如下：

    ```yaml
    port: 3339 # 服务端监听端口，默认3339
    pool_size: 20  # 代理连接池大小，默认20
    connect_timeout: 60 # 连接超时时间，默认60秒
    users:
    - username: admin
      password: "123456"
    log:
      enable: false
      path: logs

    ```



### 客户端

运行客户端：
```bash
./tunnel-client
```
以交互方式输入服务器地址、用户名、密码和本地端口。

这会将配置缓存在客户端可执行文件所在目录下的 `tunnel-client-config.yaml` 文件中。

## 星历史

[![Star History Chart](https://api.star-history.com/svg?repos=a327958099/go-tunnel&type=Date)](https://www.star-history.com/#a327958099/go-tunnel&Date)

## 许可证

本项目采用 MIT 许可证 - 有关详细信息，请参阅 [LICENSE](LICENSE) 文件。