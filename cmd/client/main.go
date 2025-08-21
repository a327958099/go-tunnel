package main

import (
	"bufio"
	"fmt"
	"go-tunnel/internal/client"
	"go-tunnel/pkg/config"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	configFile := "tunnel-client-config.yaml"

	// Check if the configuration file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// The configuration file does not exist, please enter the connection parameters
		fmt.Println("The configuration file does not exist, please enter the connection parameters：")
		cfg := interactiveInput()

		// Save the configuration file
		if err := config.SaveConfig(configFile, cfg); err != nil {
			log.Printf("Failed to save configuration file: %v", err)
		} else {
			fmt.Printf("Configuration has been saved to %s\n", configFile)
		}

		// Start the client
		client := client.NewClient(cfg)
		if err := client.Start(); err != nil {
			log.Fatalf("Failed to start client: %v", err)
		}
	} else {
		// The configuration file exists, please read the configuration file
		fmt.Printf("Read configuration file: %s\n", configFile)
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Start the client
		client := client.NewClient(cfg)
		if err := client.Start(); err != nil {
			log.Fatalf("Failed to start client: %v", err)
		}
	}
}

func interactiveInput() *config.Config {
	reader := bufio.NewReader(os.Stdin)
	cfg := &config.Config{}

	// Input the server address
	fmt.Print("Please enter the server address (default: 127.0.0.1:3339): ")
	serverAddr, _ := reader.ReadString('\n')
	serverAddr = strings.TrimSpace(serverAddr)
	if serverAddr == "" {
		serverAddr = "127.0.0.1:3339"
	}
	cfg.ServerAddr = serverAddr

	// Input the local port
	fmt.Print("Please enter the local port (default: 8000): ")
	localPortStr, _ := reader.ReadString('\n')
	localPortStr = strings.TrimSpace(localPortStr)
	if localPortStr == "" {
		cfg.LocalPort = 8000
	} else {
		if port, err := strconv.Atoi(localPortStr); err == nil {
			cfg.LocalPort = port
		} else {
			fmt.Println("The port format is incorrect, using the default value 8000")
			cfg.LocalPort = 8000
		}
	}

	// Input the connection timeout time
	fmt.Print("Please enter the connection timeout time/seconds (default: 30): ")
	timeoutStr, _ := reader.ReadString('\n')
	timeoutStr = strings.TrimSpace(timeoutStr)
	if timeoutStr == "" {
		cfg.ConnectTimeout = 30
	} else {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.ConnectTimeout = timeout
		} else {
			fmt.Println("The timeout time format is incorrect, using the default value 30")
			cfg.ConnectTimeout = 30
		}
	}

	// Input the username
	fmt.Print("Please enter the username (default: admin): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}
	cfg.Username = username

	// Input the password
	fmt.Print("Please enter the password (default: 123456): ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if password == "" {
		password = "123456"
	}
	cfg.Password = password

	return cfg
}
