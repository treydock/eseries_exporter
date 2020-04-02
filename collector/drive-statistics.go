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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

type DriveStatistics struct {
	ID                   string  `json:"diskId"`
	AverageReadOpSize    float64 `json:"averageReadOpSize"`
	AverageWriteOpSize   float64 `json:"averageWriteOpSize"`
	CombinedIOps         float64 `json:"combinedIOps"`
	CombinedResponseTime float64 `json:"combinedResponseTime"`
	CombinedThroughput   float64 `json:"combinedThroughput"`
	ReadIOps             float64 `json:"readIOps"`
	ReadOps              float64 `json:"readOps"`
	ReadPhysicalIOps     float64 `json:"readPhysicalIOps"`
	ReadResponseTime     float64 `json:"readResponseTime"`
	ReadThroughput       float64 `json:"readThroughput"`
	WriteIOps            float64 `json:"writeIOps"`
	WriteOps             float64 `json:"writeOps"`
	WritePhysicalIOps    float64 `json:"writePhysicalIOps"`
	WriteResponseTime    float64 `json:"writeResponseTime"`
	WriteThroughput      float64 `json:"writeThroughput"`
}

type DriveStatisticsCollector struct {
	AverageReadOpSize    *prometheus.Desc
	AverageWriteOpSize   *prometheus.Desc
	CombinedIOps         *prometheus.Desc
	CombinedResponseTime *prometheus.Desc
	CombinedThroughput   *prometheus.Desc
	ReadIOps             *prometheus.Desc
	ReadOps              *prometheus.Desc
	ReadPhysicalIOps     *prometheus.Desc
	ReadResponseTime     *prometheus.Desc
	ReadThroughput       *prometheus.Desc
	WriteIOps            *prometheus.Desc
	WriteOps             *prometheus.Desc
	WritePhysicalIOps    *prometheus.Desc
	WriteResponseTime    *prometheus.Desc
	WriteThroughput      *prometheus.Desc
	target               config.Target
	logger               log.Logger
	useCache             bool
}

func init() {
	registerCollector("drive-statistics", true, NewDriveStatisticsExporter)
}

func NewDriveStatisticsExporter(target config.Target, logger log.Logger, useCache bool) Collector {
	return &DriveStatisticsCollector{
		AverageReadOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "average_read_op_size_bytes"),
			"Drive statistic averageReadOpSize", []string{"tray", "slot"}, nil),
		AverageWriteOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "average_write_op_size_bytes"),
			"Drive statistic averageWriteOpSize", []string{"tray", "slot"}, nil),
		CombinedIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "combined_iops"),
			"Drive statistic combinedIOps", []string{"tray", "slot"}, nil),
		CombinedResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "combined_response_time_milliseconds"),
			"Drive statistic combinedResponseTime", []string{"tray", "slot"}, nil),
		CombinedThroughput: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "combined_throughput_mb_per_second"),
			"Drive statistic combinedThroughput", []string{"tray", "slot"}, nil),
		ReadIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_iops"),
			"Drive statistic readIOps", []string{"tray", "slot"}, nil),
		ReadOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_ops"),
			"Drive statistic readOps", []string{"tray", "slot"}, nil),
		ReadPhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_physical_iops"),
			"Drive statistic readPhysicalIOps", []string{"tray", "slot"}, nil),
		ReadResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_response_time_milliseconds"),
			"Drive statistic readResponseTime", []string{"tray", "slot"}, nil),
		ReadThroughput: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_throughput_mb_per_second"),
			"Drive statistic combinedThroughput", []string{"tray", "slot"}, nil),
		WriteIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_iops"),
			"Drive statistic writeIOps", []string{"tray", "slot"}, nil),
		WriteOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_ops"),
			"Drive statistic writeOps", []string{"tray", "slot"}, nil),
		WritePhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_physical_iops"),
			"Drive statistic writePhysicalIOps", []string{"tray", "slot"}, nil),
		WriteResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_response_time_milliseconds"),
			"Drive statistic writeResponseTime", []string{"tray", "slot"}, nil),
		WriteThroughput: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_throughput_mb_per_second"),
			"Drive statistic combinedThroughput", []string{"tray", "slot"}, nil),
		target:   target,
		logger:   logger,
		useCache: useCache,
	}
}

func (c *DriveStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.AverageReadOpSize
	ch <- c.AverageWriteOpSize
	ch <- c.CombinedIOps
	ch <- c.CombinedResponseTime
	ch <- c.CombinedThroughput
	ch <- c.ReadIOps
	ch <- c.ReadOps
	ch <- c.ReadPhysicalIOps
	ch <- c.ReadResponseTime
	ch <- c.ReadThroughput
	ch <- c.WriteIOps
	ch <- c.WriteOps
	ch <- c.WritePhysicalIOps
	ch <- c.WriteResponseTime
	ch <- c.WriteThroughput
}

func (c *DriveStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting drive-statistics metrics")
	collectTime := time.Now()
	var errorMetric int
	inventory, statistics, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	trays := make(map[string]int)
	drives := make(map[string]Drive)
	for _, t := range inventory.Trays {
		trays[t.TrayRef] = t.ID
	}
	for _, d := range inventory.Drives {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		drives[d.ID] = d
	}

	for _, s := range statistics {
		drive, ok := drives[s.ID]
		if !ok {
			drive = Drive{Slot: s.ID}
		}
		ch <- prometheus.MustNewConstMetric(c.AverageReadOpSize, prometheus.GaugeValue, s.AverageReadOpSize, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.AverageWriteOpSize, prometheus.GaugeValue, s.AverageWriteOpSize, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.CombinedIOps, prometheus.GaugeValue, s.CombinedIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.CombinedResponseTime, prometheus.GaugeValue, s.CombinedResponseTime, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.CombinedThroughput, prometheus.GaugeValue, s.CombinedThroughput, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadIOps, prometheus.GaugeValue, s.ReadIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadOps, prometheus.GaugeValue, s.ReadOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadPhysicalIOps, prometheus.GaugeValue, s.ReadPhysicalIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadResponseTime, prometheus.GaugeValue, s.ReadResponseTime, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadThroughput, prometheus.GaugeValue, s.ReadThroughput, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteIOps, prometheus.GaugeValue, s.WriteIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteOps, prometheus.GaugeValue, s.WriteOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WritePhysicalIOps, prometheus.GaugeValue, s.WritePhysicalIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteResponseTime, prometheus.GaugeValue, s.WriteResponseTime, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteThroughput, prometheus.GaugeValue, s.WriteThroughput, drive.TrayID, drive.Slot)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "drive-statistics")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "drive-statistics")
}

func (c *DriveStatisticsCollector) collect() (DrivesInventory, []DriveStatistics, error) {
	var inventory DrivesInventory
	var statistics []DriveStatistics
	inventoryBody, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/hardware-inventory", c.target.Name), c.logger)
	if err != nil {
		return inventory, nil, err
	}
	err = json.Unmarshal(inventoryBody, &inventory)
	if err != nil {
		return inventory, nil, err
	}
	if len(inventory.Drives) == 0 {
		return inventory, nil, fmt.Errorf("No drives returned")
	}
	statisticsBody, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/analysed-drive-statistics", c.target.Name), c.logger)
	if err != nil {
		return inventory, nil, err
	}
	err = json.Unmarshal(statisticsBody, &statistics)
	if err != nil {
		return inventory, nil, err
	}
	return inventory, statistics, nil
}
