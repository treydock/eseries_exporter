// Copyright 2020 Trey Dockendorf
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Modules map[string]*Module `yaml:"modules"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

type Module struct {
	User       string   `yaml:"user"`
	Password   string   `yaml:"password"`
	ProxyURL   string   `yaml:"proxy_url"`
	Collectors []string `yaml:"collectors"`
	Timeout    int      `yaml:"timeout"`
}

type Target struct {
	Name       string
	User       string
	Password   string
	ProxyURL   string
	Collectors []string
	BaseURL    *url.URL
	HttpClient *http.Client
}

func (sc *SafeConfig) ReloadConfig(configFile string) error {
	var c = &Config{}
	yamlReader, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("Error reading config file %s: %s", configFile, err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)
	if err := decoder.Decode(c); err != nil {
		return fmt.Errorf("Error parsing config file %s: %s", configFile, err)
	}
	for key := range c.Modules {
		module := c.Modules[key]
		if module.Timeout == 0 {
			module.Timeout = 10
		}
		if module.ProxyURL == "" {
			return fmt.Errorf("Module %s must define 'proxy_url' value", key)
		}
		if module.User == "" {
			return fmt.Errorf("Module %s must define 'user' value", key)
		}
		if module.Password == "" {
			return fmt.Errorf("Module %s must define 'password' value", key)
		}
		c.Modules[key] = module
	}
	sc.Lock()
	sc.C = c
	sc.Unlock()
	return nil
}
