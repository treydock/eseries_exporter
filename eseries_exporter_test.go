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

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/treydock/eseries_exporter/config"
)

const (
	address = "localhost:19313"
)

func SetupServer() *config.Config {
	fixtureData, err := os.ReadFile("collector/testdata/drives.json")
	if err != nil {
		fmt.Printf("Error loading fixture data: %s", err.Error())
		os.Exit(1)
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
	}))
	sslServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
	}))
	module := &config.Module{
		User:       "test",
		Password:   "test",
		Collectors: []string{"drives"},
		ProxyURL:   server.URL,
	}
	sslModule := &config.Module{
		User:        "test",
		Password:    "test",
		Collectors:  []string{"drives"},
		ProxyURL:    sslServer.URL,
		RootCA:      "collector/testdata/rootCA.crt",
		InsecureSSL: true,
	}
	c := &config.Config{}
	c.Modules = make(map[string]*config.Module)
	c.Modules["default"] = module
	c.Modules["ssl"] = sslModule
	return c
}

func TestMetricsHandler(t *testing.T) {
	c := SetupServer()
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	go func() {
		http.Handle("/eseries", metricsHandler(c, logger))
		err := http.ListenAndServe(address, nil)
		if err != nil {
			os.Exit(1)
		}
	}()
	body, err := queryExporter("target=test1", http.StatusOK)
	if err != nil {
		t.Fatalf("Unexpected error GET /eseries: %s", err.Error())
	}
	if !strings.Contains(body, "eseries_exporter_collect_error{collector=\"drives\"} 0") {
		t.Errorf("Unexpected value for eseries_exporter_collect_error")
	}

	body, err = queryExporter("target=test1&module=ssl", http.StatusOK)
	if err != nil {
		t.Fatalf("Unexpected error GET /eseries: %s", err.Error())
	}
	if !strings.Contains(body, "eseries_exporter_collect_error{collector=\"drives\"} 0") {
		t.Errorf("Unexpected value for eseries_exporter_collect_error")
	}

	_, _ = queryExporter("", http.StatusBadRequest)

	_, _ = queryExporter("module=dne", http.StatusNotFound)
}

func queryExporter(param string, want int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/eseries?%s", address, param))
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := resp.Body.Close(); err != nil {
		return "", err
	}
	if have := resp.StatusCode; want != have {
		return "", fmt.Errorf("want /eseries status code %d, have %d. Body:\n%s", want, have, b)
	}
	return string(b), nil
}
