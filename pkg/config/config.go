package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddr     string `yaml:"server_addr"`
	Port           int    `yaml:"port"`
	LocalPort      int    `yaml:"local_port"`
	ConnectTimeout int    `yaml:"connect_timeout"`
	PoolSize       int    `yaml:"pool_size"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Users          []User `yaml:"users"`
	Log            Log    `yaml:"log"`
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Log struct {
	Enable bool   `yaml:"enable"`
	Path   string `yaml:"path"`
}

func SaveConfig(path string, config *Config) error {
	// Create a configuration structure containing only the client fields
	clientConfig := struct {
		ServerAddr     string `yaml:"server_addr"`
		LocalPort      int    `yaml:"local_port"`
		ConnectTimeout int    `yaml:"connect_timeout"`
		Username       string `yaml:"username"`
		Password       string `yaml:"password"`
	}{
		ServerAddr:     config.ServerAddr,
		LocalPort:      config.LocalPort,
		ConnectTimeout: config.ConnectTimeout,
		Username:       config.Username,
		Password:       config.Password,
	}

	data, err := yaml.Marshal(clientConfig)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// The configuration file does not exist, please enter the connection parameters
		var defaultConfig *Config
		var data []byte

		if strings.Contains(path, "client") {
			// The client default configuration
			defaultConfig = &Config{
				ServerAddr:     "127.0.0.1:3339",
				LocalPort:      8000,
				ConnectTimeout: 30,
				Username:       "admin",
				Password:       "123456",
			}
			// Create a configuration structure containing only the client fields
			clientConfig := struct {
				ServerAddr     string `yaml:"server_addr"`
				LocalPort      int    `yaml:"local_port"`
				ConnectTimeout int    `yaml:"connect_timeout"`
				Username       string `yaml:"username"`
				Password       string `yaml:"password"`
			}{
				ServerAddr:     defaultConfig.ServerAddr,
				LocalPort:      defaultConfig.LocalPort,
				ConnectTimeout: defaultConfig.ConnectTimeout,
				Username:       defaultConfig.Username,
				Password:       defaultConfig.Password,
			}
			var err error
			data, err = yaml.Marshal(clientConfig)
			if err != nil {
				return nil, err
			}
		} else {
			// The server default configuration
			defaultConfig = &Config{
				Port:           3339,
				PoolSize:       20,
				ConnectTimeout: 60,
				Users: []User{
					{Username: "admin", Password: "123456"},
				},
				Log: Log{
					Enable: false,
					Path:   "logs",
				},
			}
			// Create a configuration structure containing only the server fields
			serverConfig := struct {
				Port           int    `yaml:"port"`
				PoolSize       int    `yaml:"pool_size"`
				ConnectTimeout int    `yaml:"connect_timeout"`
				Users          []User `yaml:"users"`
				Log            Log    `yaml:"log"`
			}{
				Port:           defaultConfig.Port,
				PoolSize:       defaultConfig.PoolSize,
				ConnectTimeout: defaultConfig.ConnectTimeout,
				Users:          defaultConfig.Users,
				Log:            defaultConfig.Log,
			}
			var err error
			data, err = yaml.Marshal(serverConfig)
			if err != nil {
				return nil, err
			}
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			return nil, err
		}
		return defaultConfig, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
