package client

import (
	"bufio"
	"fmt"
	"go-tunnel/pkg/config"
	"go-tunnel/pkg/logger"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// Client represents a tunnel client

type Client struct {
	config   *config.Config
	stopChan chan struct{}
	stopped  bool
	mu       sync.Mutex
}

func NewClient(config *config.Config) *Client {
	return &Client{
		config:   config,
		stopChan: make(chan struct{}),
	}
}

func (c *Client) Start() error {
	logger.InfoLogger.Println("Client started")
	logger.InfoLogger.Println("Press Ctrl+C to stop the client")

	for {
		// Reset stop channel and flag for each connection attempt
		c.mu.Lock()
		c.stopChan = make(chan struct{})
		c.stopped = false
		c.mu.Unlock()
		c.runControlConnection()
		logger.InfoLogger.Println("Disconnected from server, reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Client) runControlConnection() {
	// Check if ServerAddr already contains a port
	var serverAddr string
	if c.config.ServerAddr != "" {
		serverAddr = c.config.ServerAddr
	} else {
		serverAddr = fmt.Sprintf("localhost:%d", c.config.Port)
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()

	if c.config.Username == "" || c.config.Password == "" {
		logger.ErrorLogger.Println("Username or password not configured for the client.")
		return
	}

	// Authenticate
	authStr := fmt.Sprintf("AUTH %s:%s\n", c.config.Username, c.config.Password)
	_, err = conn.Write([]byte(authStr))
	if err != nil {
		logger.ErrorLogger.Printf("Failed to send auth info: %v", err)
		return
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		logger.ErrorLogger.Printf("Failed to read auth response: %v", err)
		return
	}

	// Check if authentication is successful
	if !strings.HasPrefix(line, "AUTH_OK") {
		logger.ErrorLogger.Printf("Server authentication failed: %s", strings.TrimSpace(line))
		return
	}

	// Parse the connection pool size
	poolSize := 10 // Default value
	if strings.Contains(line, "pool_size=") {
		parts := strings.Split(line, "pool_size=")
		if len(parts) > 1 {
			poolSizeStr := strings.TrimSpace(strings.Split(parts[1], "\n")[0])
			if parsedSize, parseErr := fmt.Sscanf(poolSizeStr, "%d", &poolSize); parseErr == nil && parsedSize == 1 {
				logger.InfoLogger.Printf("Server specified pool_size: %d", poolSize)
			} else {
				logger.ErrorLogger.Printf("Failed to parse pool_size, using default: %d", poolSize)
			}
		}
	}

	logger.InfoLogger.Printf("Successfully authenticated with server, creating %d proxy connections.", poolSize)

	// Start proxy connections with server-specified pool size
	for i := 0; i < poolSize; i++ {
		go c.startProxyConnection(c.config.Username)
	}

	// Keep control connection alive
	for {
		_, err := conn.Read(make([]byte, 1)) // Dummy read
		if err != nil {
			logger.ErrorLogger.Printf("Control connection lost: %v", err)
			// Signal all proxy connections to stop
			c.mu.Lock()
			if !c.stopped {
				c.stopped = true
				close(c.stopChan)
			}
			c.mu.Unlock()
			return
		}
	}
}

func (c *Client) startProxyConnection(username string) {
	for {
		// Check if we should stop
		select {
		case <-c.stopChan:
			logger.InfoLogger.Println("Proxy connection stopped due to control connection loss")
			return
		default:
		}

		// Check if ServerAddr already contains a port
		var serverAddr string
		if c.config.ServerAddr != "" {
			serverAddr = c.config.ServerAddr
		} else {
			serverAddr = fmt.Sprintf("localhost:%d", c.config.Port)
		}

		proxyConn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create proxy connection: %v", err)
			// Wait longer before retrying to avoid spamming
			select {
			case <-c.stopChan:
				return
			case <-time.After(10 * time.Second):
				continue
			}
		}

		proxyStr := fmt.Sprintf("PROXY %s\n", username)
		_, err = proxyConn.Write([]byte(proxyStr))
		if err != nil {
			logger.ErrorLogger.Printf("Failed to register proxy connection: %v", err)
			proxyConn.Close()
			// Wait before retrying
			select {
			case <-c.stopChan:
				return
			case <-time.After(10 * time.Second):
				continue
			}
		}

		logger.InfoLogger.Println("Proxy connection established, waiting for server to use it.")

		// Wait for data from the server, then connect to local service
		buf := make([]byte, 1)
		_, err = proxyConn.Read(buf)
		if err != nil {
			if err != io.EOF {
				logger.ErrorLogger.Printf("Proxy connection error: %v", err)
			}
			proxyConn.Close()
			// Wait before retrying
			select {
			case <-c.stopChan:
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		localConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", c.config.LocalPort))
		if err != nil {
			logger.ErrorLogger.Printf("Failed to connect to local service: %v", err)
			proxyConn.Close()
			time.Sleep(1 * time.Second) // Wait a short time before retrying
			continue
		}

		// Write the first byte back
		localConn.Write(buf)

		// Bridge the connections
		go io.Copy(localConn, proxyConn)
		io.Copy(proxyConn, localConn)
		proxyConn.Close()
		localConn.Close()
		logger.InfoLogger.Println("Finished proxying a connection, creating new proxy connection.")
		// After connection handling, the loop will automatically create a new proxy connection
	}
}
