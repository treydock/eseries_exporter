// MIT License
//
// Copyright (c) 2020 Ohio Supercomputer Center
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package collector

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/treydock/eseries_exporter/config"
)

func TestHardwareInventoryCollector(t *testing.T) {
	fixtureData, err := os.ReadFile("testdata/hardware-inventory.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `
	# HELP eseries_battery_status Status of battery hardware device
	# TYPE eseries_battery_status gauge
	eseries_battery_status{slot="1",status="configMismatch",tray="99"} 0
	eseries_battery_status{slot="1",status="expired",tray="99"} 0
	eseries_battery_status{slot="1",status="failed",tray="99"} 0
	eseries_battery_status{slot="1",status="fullCharging",tray="99"} 0
	eseries_battery_status{slot="1",status="learning",tray="99"} 0
	eseries_battery_status{slot="1",status="maintenanceCharging",tray="99"} 0
	eseries_battery_status{slot="1",status="nearExpiration",tray="99"} 0
	eseries_battery_status{slot="1",status="notInConfig",tray="99"} 0
	eseries_battery_status{slot="1",status="optimal",tray="99"} 1
	eseries_battery_status{slot="1",status="overtemp",tray="99"} 0
	eseries_battery_status{slot="1",status="removed",tray="99"} 0
	eseries_battery_status{slot="1",status="replacementRequired",tray="99"} 0
	eseries_battery_status{slot="1",status="unknown",tray="99"} 0
	eseries_battery_status{slot="2",status="configMismatch",tray="99"} 0
	eseries_battery_status{slot="2",status="expired",tray="99"} 0
	eseries_battery_status{slot="2",status="failed",tray="99"} 0
	eseries_battery_status{slot="2",status="fullCharging",tray="99"} 0
	eseries_battery_status{slot="2",status="learning",tray="99"} 0
	eseries_battery_status{slot="2",status="maintenanceCharging",tray="99"} 0
	eseries_battery_status{slot="2",status="nearExpiration",tray="99"} 0
	eseries_battery_status{slot="2",status="notInConfig",tray="99"} 0
	eseries_battery_status{slot="2",status="optimal",tray="99"} 0
	eseries_battery_status{slot="2",status="overtemp",tray="99"} 0
	eseries_battery_status{slot="2",status="removed",tray="99"} 0
	eseries_battery_status{slot="2",status="replacementRequired",tray="99"} 0
	eseries_battery_status{slot="2",status="unknown",tray="99"} 1
	# HELP eseries_cache_memory_dimm_status Status of cache memory DIMM hardware device
	# TYPE eseries_cache_memory_dimm_status gauge
	eseries_cache_memory_dimm_status{slot="1",status="empty",tray="99"} 0
	eseries_cache_memory_dimm_status{slot="1",status="failed",tray="99"} 0
	eseries_cache_memory_dimm_status{slot="1",status="optimal",tray="99"} 1
	eseries_cache_memory_dimm_status{slot="1",status="unknown",tray="99"} 0
	eseries_cache_memory_dimm_status{slot="2",status="empty",tray="99"} 0
	eseries_cache_memory_dimm_status{slot="2",status="failed",tray="99"} 0
	eseries_cache_memory_dimm_status{slot="2",status="optimal",tray="99"} 0
	eseries_cache_memory_dimm_status{slot="2",status="unknown",tray="99"} 1
	# HELP eseries_fan_status Status of fan hardware device
	# TYPE eseries_fan_status gauge
	eseries_fan_status{slot="1",status="failed",tray="99"} 0
	eseries_fan_status{slot="1",status="optimal",tray="99"} 1
	eseries_fan_status{slot="1",status="removed",tray="99"} 0
	eseries_fan_status{slot="1",status="unknown",tray="99"} 0
	eseries_fan_status{slot="2",status="failed",tray="99"} 0
	eseries_fan_status{slot="2",status="optimal",tray="99"} 0
	eseries_fan_status{slot="2",status="removed",tray="99"} 0
	eseries_fan_status{slot="2",status="unknown",tray="99"} 1
	# HELP eseries_power_supply_status Status of power supply hardware device
	# TYPE eseries_power_supply_status gauge
	eseries_power_supply_status{slot="1",status="failed",tray="99"} 0
	eseries_power_supply_status{slot="1",status="noinput",tray="99"} 0
	eseries_power_supply_status{slot="1",status="optimal",tray="99"} 1
	eseries_power_supply_status{slot="1",status="removed",tray="99"} 0
	eseries_power_supply_status{slot="1",status="unknown",tray="99"} 0
	eseries_power_supply_status{slot="2",status="failed",tray="99"} 0
	eseries_power_supply_status{slot="2",status="noinput",tray="99"} 0
	eseries_power_supply_status{slot="2",status="optimal",tray="99"} 0
	eseries_power_supply_status{slot="2",status="removed",tray="99"} 0
	eseries_power_supply_status{slot="2",status="unknown",tray="99"} 1
	# HELP eseries_thermal_sensor_status Status of thermal sensor hardware device
	# TYPE eseries_thermal_sensor_status gauge
	eseries_thermal_sensor_status{slot="1",status="maxTempExceed",tray="99"} 0
	eseries_thermal_sensor_status{slot="1",status="nominalTempExceed",tray="99"} 0
	eseries_thermal_sensor_status{slot="1",status="optimal",tray="99"} 1
	eseries_thermal_sensor_status{slot="1",status="removed",tray="99"} 0
	eseries_thermal_sensor_status{slot="1",status="unknown",tray="99"} 0
	eseries_thermal_sensor_status{slot="2",status="maxTempExceed",tray="99"} 0
	eseries_thermal_sensor_status{slot="2",status="nominalTempExceed",tray="99"} 0
	eseries_thermal_sensor_status{slot="2",status="optimal",tray="99"} 0
	eseries_thermal_sensor_status{slot="2",status="removed",tray="99"} 0
	eseries_thermal_sensor_status{slot="2",status="unknown",tray="99"} 1
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="hardware-inventory"} 0
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test",
		User:       "test",
		Password:   "test",
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewHardwareInventoryExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 64 {
		t.Errorf("Unexpected collection count %d, expected 64", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_battery_status", "eseries_fan_status",
		"eseries_power_supply_status", "eseries_cache_memory_dimm_status",
		"eseries_thermal_sensor_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestHardwareInventoryCollectorError(t *testing.T) {
	expected := `
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="hardware-inventory"} 1
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "error", http.StatusNotFound)
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test",
		User:       "test",
		Password:   "test",
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewHardwareInventoryExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_battery_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
