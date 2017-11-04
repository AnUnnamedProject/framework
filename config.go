// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Current framework config parameters:
// author         string
// cache          string
// cache_config   string
// compress_html  bool
// compress_css   bool
// compress_js    bool
// database       string
// database_conn  string
// mode           string
// name           string
// port           int
// grpc_port      int
// grpc_cert      string
// grpc_cert_key  string
// session        string
// session_config JSON
// smtp_server    string
// smtp_auth      string
// smtp_username  string
// smtp_password  string
// version        string
// template_left  string
// template_right string
// pprof          string
type (
	// Config struct {
	// 	Author        string `json:"author"`
	// 	Cache         string `json:"cache"`
	// 	CacheConfig   string `json:"cache"`
	// 	CompressHTML  bool   `json:"compress_html"`
	// 	CompressCSS   bool   `json:"compress_css"`
	// 	CompressJS    bool   `json:"compress_js"`
	// 	Database      string `json:"database"`
	// 	DatabaseConn  string `json:"database_conn"`
	// 	Mode          string `json:"mode"`
	// 	Name          string `json:"name"`
	// 	Port          int    `json:"port"`
	// 	Session       string `json:"session"`
	// 	SessionConfig JSON   `json:"session_config"`
	// 	SMTPServer    string `json:"smtp_server"`
	// 	SMTPAuth      string `json:"smtp_auth"`
	// 	SMTPUsername  string `json:"smtp_username"`
	// 	SMTPPassword  string `json:"smtp_password"`
	// 	Version       string `json:"version"`
	// 	TemplateLeft  string `json:"template_left"`
	// 	TemplateRight string `json:"template_right"`
	// 	Pprof         string `json:"pprof"`
	// }

	// Configuration is the interface for app configuration.
	Configuration interface {
		Set(key string, value interface{})
		Get(key string) interface{}
		String(key string) string
		Int(key string) int
		Bool(key string) bool
	}

	// ConfigData contains the configuration data
	ConfigData struct {
		data JSON
	}
)

// Config is the framework global configuration
var Config Configuration

// DefaultConfig initialize the Config with basic configuration settings.
func DefaultConfig() *ConfigData {
	var data JSON

	data["name"] = "AnUnnamedApp"
	data["author"] = "AnUnnamedProject"
	data["port"] = float64(8080)
	data["mode"] = "debug"
	data["compress_html"] = true
	data["compress_css"] = true
	data["compress_js"] = true
	data["template_left"] = "{{"
	data["template_right"] = "}}"

	return &ConfigData{data: data}
}

// LoadConfig loads the configuration file from specified path and returns the Config object.
func LoadConfig(path string) Configuration {
	config := DefaultConfig()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("unable to load configuration", err)
		return config
	}

	if err := json.Unmarshal(data, &config.data); err != nil {
		fmt.Println("unable to decode json", err)
	}

	return config
}

// Set updates a key with value.
func (c *ConfigData) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get return config's value by specified key.
func (c *ConfigData) Get(key string) interface{} {
	return c.data[key]
}

// String returns the config's value by specified key, cast to string if possible.
// If data type don't match, return an empty string
func (c *ConfigData) String(key string) string {
	v := c.Get(key)

	if _, ok := v.(string); !ok {
		return ""
	}

	return v.(string)
}

// Int returns the config's value by specified key, cast to int if possible.
// If data type don't match, return 0
func (c *ConfigData) Int(key string) int {
	v := c.Get(key)

	if _, ok := v.(int); !ok {
		return 0
	}

	return v.(int)
}

// Bool returns the config's value by specified key, cast to bool if possible.
// If data type don't match, return false
func (c *ConfigData) Bool(key string) bool {
	v := c.Get(key)

	if _, ok := v.(bool); !ok {
		return false
	}

	return v.(bool)

}
