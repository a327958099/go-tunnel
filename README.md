# Go Tunnel

[简体中文](README_zh.md)

A extremely lightweight reverse proxy tool implemented in Go, specifically designed for seamlessly forwarding external network requests to local services during the development phase, used for receiving payment callbacks, notification messages, and other scenarios.

## Features

- Supports http, https, ws, wss protocols.
- Easy to configure with a single `tunnel-server-config.yaml` file.
- Secure client-server communication with username/password authentication.
- No runtime dependencies, the program can be directly executed without any additional software installation.

## Architecture

### Core Communication Mechanism

Go Tunnel uses a transparent TCP data bridging architecture that supports any TCP-based protocol without protocol-specific handling:

```
External Client ↔ Server ↔ Proxy Connection ↔ Client ↔ Local Service
```

**Key Components:**

1. **Control Connection**: Handles authentication and management between client and server
2. **Proxy Connection Pool**: Pre-established connections (default 20) waiting for incoming requests
3. **Transparent Data Bridging**: Uses `io.Copy()` for zero-overhead byte-level data forwarding

**Data Flow Process:**

1. **External Request**: Server receives incoming connection from external client
2. **Connection Assignment**: Server assigns an available proxy connection from the pool
3. **Transparent Forwarding**: Server bridges external connection ↔ proxy connection using `io.Copy()`
4. **Local Forwarding**: Client bridges proxy connection ↔ local service using `io.Copy()`

**Advantages:**

- **Protocol Agnostic**: Supports HTTP/HTTPS/WebSocket and any TCP-based protocol
- **Zero Overhead**: Direct byte stream forwarding without additional processing
- **High Performance**: Leverages Go's efficient I/O copying mechanisms
- **Transparent**: Maintains original protocol integrity without modification

## Supported Platforms

- Windows
- macOS
- Linux

## Direct Download

Download links: [https://github.com/a327958099/go-tunnel/releases](https://github.com/a327958099/go-tunnel/releases)

## Build from Source

1.  Clone the repository:
    ```bash
    git clone https://github.com/a327958099/go-tunnel.git
    cd go-tunnel
    ```

2.  Build the server and client:
    ```bash
    go build -o tunnel-server ./cmd/server
    go build -o tunnel-client ./cmd/client
    ```
    This will create `tunnel-server` and `tunnel-client` executables in the project root directory.

## Usage

### Server

1.  Run the server:

    ```bash
    ./tunnel-server
    ```

2.  In the same directory as the server executable, a `tunnel-server-config.yaml` file will be automatically created with the following content:

    ```yaml
    port: 3339  # server listen port, default 3339
    pool_size: 20  # proxy connection pool size, default 20
    connect_timeout: 60 # connection timeout, default 60 seconds
    users:
      - username: admin
        password: 123456
    log:
      enable: true
      path: logs
    ```



### Client

Run the client:

```bash
./tunnel-client
```

Enter the server address, username, password, and local port interactively.

This will cache the configuration in a `tunnel-client-config.yaml` file in the same directory as the client executable.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=a327958099/go-tunnel&type=Date)](https://www.star-history.com/#a327958099/go-tunnel&Date)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
