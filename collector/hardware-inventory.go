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

package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

var (
	batteryStatuses = []string{"optimal", "fullCharging", "nearExpiration", "failed", "removed", "notInConfig",
		"configMismatch", "learning", "overtemp", "expired", "maintenanceCharging", "replacementRequired"}
	fanStatuses             = []string{"optimal", "failed", "removed"}
	powerSupplyStatuses     = []string{"optimal", "failed", "removed", "noinput"}
	cacheMemoryDimmStatuses = []string{"optimal", "empty", "failed"}
	thermalSensorStatuses   = []string{"optimal", "nominalTempExceed", "maxTempExceed", "removed"}
)

type HardwareInventory struct {
	Trays            []Tray            `json:"trays"`
	Batteries        []Battery         `json:"batteries"`
	Fans             []Fan             `json:"fans"`
	PowerSupplies    []PowerSupply     `json:"powerSupplies"`
	CacheMemoryDimms []CacheMemoryDimm `json:"cacheMemoryDimms"`
	ThermalSensors   []ThermalSensor   `json:"thermalSensors"`
}

type Battery struct {
	ID               string `json:"id"`
	TrayID           string
	Slot             string
	Status           string           `json:"status"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type Fan struct {
	ID               string `json:"id"`
	TrayID           string
	Slot             string
	Status           string           `json:"status"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PowerSupply struct {
	ID               string `json:"id"`
	TrayID           string
	Slot             string
	Status           string           `json:"status"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type CacheMemoryDimm struct {
	TrayID           string
	Slot             string
	Status           string           `json:"status"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type ThermalSensor struct {
	ID               string `json:"id"`
	TrayID           string
	Slot             string
	Status           string           `json:"status"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	Slot    int    `json:"slot"`
	TrayRef string `json:"trayRef"`
}

type Tray struct {
	TrayRef string `json:"trayRef"`
	ID      int    `json:"trayId"`
}

type HardwareInventoryCollector struct {
	BatteryStatus         *prometheus.Desc
	FanStatus             *prometheus.Desc
	PowerSupplyStatus     *prometheus.Desc
	CacheMemoryDimmStatus *prometheus.Desc
	ThermalSensorStatus   *prometheus.Desc
	target                config.Target
	logger                log.Logger
}

func init() {
	registerCollector("hardware-inventory", true, NewHardwareInventoryExporter)
}

func NewHardwareInventoryExporter(target config.Target, logger log.Logger) Collector {
	return &HardwareInventoryCollector{
		BatteryStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "battery", "status"),
			"Status of battery hardware device", []string{"tray", "slot", "status"}, nil),
		FanStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "fan", "status"),
			"Status of fan hardware device", []string{"tray", "slot", "status"}, nil),
		PowerSupplyStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "power_supply", "status"),
			"Status of power supply hardware device", []string{"tray", "slot", "status"}, nil),
		CacheMemoryDimmStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "cache_memory_dimm", "status"),
			"Status of cache memory DIMM hardware device", []string{"tray", "slot", "status"}, nil),
		ThermalSensorStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "thermal_sensor", "status"),
			"Status of thermal sensor hardware device", []string{"tray", "slot", "status"}, nil),
		target: target,
		logger: logger,
	}
}

func (c *HardwareInventoryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.BatteryStatus
	ch <- c.FanStatus
	ch <- c.PowerSupplyStatus
	ch <- c.CacheMemoryDimmStatus
	ch <- c.ThermalSensorStatus
}

func (c *HardwareInventoryCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting hardware-inventory metrics")
	collectTime := time.Now()
	var errorMetric int
	inventory, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	trays := make(map[string]int)
	for _, t := range inventory.Trays {
		trays[t.TrayRef] = t.ID
	}
	for _, d := range inventory.Batteries {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		for _, s := range batteryStatuses {
			var value float64
			if strings.EqualFold(s, d.Status) {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.BatteryStatus, prometheus.GaugeValue, value, d.TrayID, d.Slot, s)
		}
		var unknown float64
		if !sliceContains(batteryStatuses, d.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.BatteryStatus, prometheus.GaugeValue, unknown, d.TrayID, d.Slot, "unknown")
	}
	for _, d := range inventory.Fans {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		for _, s := range fanStatuses {
			var value float64
			if strings.EqualFold(s, d.Status) {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.FanStatus, prometheus.GaugeValue, value, d.TrayID, d.Slot, s)
		}
		var unknown float64
		if !sliceContains(fanStatuses, d.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.FanStatus, prometheus.GaugeValue, unknown, d.TrayID, d.Slot, "unknown")
	}
	for _, d := range inventory.PowerSupplies {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		for _, s := range powerSupplyStatuses {
			var value float64
			if strings.EqualFold(s, d.Status) {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.PowerSupplyStatus, prometheus.GaugeValue, value, d.TrayID, d.Slot, s)
		}
		var unknown float64
		if !sliceContains(powerSupplyStatuses, d.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.PowerSupplyStatus, prometheus.GaugeValue, unknown, d.TrayID, d.Slot, "unknown")
	}
	for _, d := range inventory.CacheMemoryDimms {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		for _, s := range cacheMemoryDimmStatuses {
			var value float64
			if strings.EqualFold(s, d.Status) {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.CacheMemoryDimmStatus, prometheus.GaugeValue, value, d.TrayID, d.Slot, s)
		}
		var unknown float64
		if !sliceContains(cacheMemoryDimmStatuses, d.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.CacheMemoryDimmStatus, prometheus.GaugeValue, unknown, d.TrayID, d.Slot, "unknown")
	}
	for _, d := range inventory.ThermalSensors {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		for _, s := range thermalSensorStatuses {
			var value float64
			if strings.EqualFold(s, d.Status) {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.ThermalSensorStatus, prometheus.GaugeValue, value, d.TrayID, d.Slot, s)
		}
		var unknown float64
		if !sliceContains(thermalSensorStatuses, d.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.ThermalSensorStatus, prometheus.GaugeValue, unknown, d.TrayID, d.Slot, "unknown")
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "hardware-inventory")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "hardware-inventory")
}

func (c *HardwareInventoryCollector) collect() (HardwareInventory, error) {
	var inventory HardwareInventory
	body, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/hardware-inventory", c.target.Name), c.logger)
	if err != nil {
		return inventory, err
	}
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return inventory, err
	}
	return inventory, nil
}
