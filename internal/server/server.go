package server

import (
	"bufio"
	"fmt"
	"go-tunnel/pkg/config"
	"go-tunnel/pkg/logger"
	"io"
	"net"
	"strings"
	"sync"
)

// Control represents a client's control connection

type Control struct {
	conn         net.Conn
	proxyConns   chan net.Conn
	username     string
	shutdown     chan struct{}
}

type Server struct {
	config   *config.Config
	controls sync.Map // map[string]*Control
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() error {
	logger.InfoLogger.Printf("Server started on port %d", s.config.Port)
	logger.InfoLogger.Println("Press Ctrl+C to stop the server")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.ErrorLogger.Printf("Failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		conn.Close()
		return
	}

	switch parts[0] {
	case "AUTH":
		s.handleAuth(conn, parts)
	case "PROXY":
		s.handleProxy(conn, parts)
	default:
		s.handlePublic(conn, line, reader)
	}
}

func (s *Server) handleAuth(conn net.Conn, parts []string) {
	if len(parts) != 2 {
		conn.Write([]byte("AUTH_FAILED invalid format\n"))
		conn.Close()
		return
	}

	authParts := strings.Split(parts[1], ":")
	if len(authParts) != 2 {
		conn.Write([]byte("AUTH_FAILED invalid auth string\n"))
		conn.Close()
		return
	}
	username, password := authParts[0], authParts[1]

	validUser := false
	for _, user := range s.config.Users {
		if user.Username == username && user.Password == password {
			validUser = true
			break
		}
	}

	if !validUser {
		conn.Write([]byte("AUTH_FAILED invalid credentials\n"))
		conn.Close()
		logger.WarningLogger.Printf("Failed auth attempt for user %s from %s", username, conn.RemoteAddr())
		return
	}

	if c, ok := s.controls.Load(username); ok {
		// Cleanup previous control connection if it exists
		close(c.(*Control).shutdown)
		c.(*Control).conn.Close()
	}

	// 获取连接池大小，默认为20
	poolSize := s.config.PoolSize
	if poolSize <= 0 {
		poolSize = 20
	}

	// 发送认证成功消息和连接池大小
	authResponse := fmt.Sprintf("AUTH_OK pool_size=%d\n", poolSize)
	conn.Write([]byte(authResponse))
	logger.InfoLogger.Printf("User %s authenticated from %s, pool_size=%d", username, conn.RemoteAddr(), poolSize)

	control := &Control{
		conn:       conn,
		proxyConns: make(chan net.Conn, poolSize),
		username:   username,
		shutdown:   make(chan struct{}),
	}
	s.controls.Store(username, control)

	// Keep connection alive, and remove it from map on disconnect
	<-control.shutdown // Wait until a new control connection is made or server stops
	s.controls.Delete(username)
	logger.InfoLogger.Printf("Control connection for %s closed.", username)
}

func (s *Server) handleProxy(conn net.Conn, parts []string) {
	if len(parts) != 2 {
		conn.Close()
		return
	}
	username := parts[1]
	if c, ok := s.controls.Load(username); ok {
		control := c.(*Control)
		select {
		case control.proxyConns <- conn:
			logger.InfoLogger.Printf("Added new proxy connection for %s from %s", username, conn.RemoteAddr())
		case <-control.shutdown:
			conn.Close()
		}
	} else {
		conn.Close()
	}
}

func (s *Server) handlePublic(conn net.Conn, firstLine string, reader *bufio.Reader) {
	// This logic assumes a single client. For multi-client, we'd need a way to route.
	// For now, we find the first (and only) control.
	var control *Control
	s.controls.Range(func(key, value interface{}) bool {
		control = value.(*Control)
		return false // stop iterating
	})

	if control == nil {
		logger.ErrorLogger.Println("No authenticated client to handle public request.")
		conn.Close()
		return
	}

	logger.InfoLogger.Printf("New public connection from %s for user %s", conn.RemoteAddr(), control.username)

	var proxyConn net.Conn
	select {
	case proxyConn = <-control.proxyConns:
		logger.InfoLogger.Printf("Found available proxy connection for %s", control.username)
	case <-control.shutdown:
		conn.Close()
		return
	}

	// Forward the first line that was already read
	proxyConn.Write([]byte(firstLine))

	// Bridge the connections
	go io.Copy(proxyConn, reader)
	io.Copy(conn, proxyConn)
	conn.Close()
	proxyConn.Close()
	logger.InfoLogger.Printf("Finished proxying for %s", conn.RemoteAddr())
}
