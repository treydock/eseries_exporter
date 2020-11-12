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
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

type AnalysedDriveStatistics struct {
	ID                   string  `json:"diskId"`
	AverageReadOpSize    float64 `json:"averageReadOpSize"`
	AverageWriteOpSize   float64 `json:"averageWriteOpSize"`
	CombinedResponseTime float64 `json:"combinedResponseTime"`
	ReadPhysicalIOps     float64 `json:"readPhysicalIOps"`
	ReadResponseTime     float64 `json:"readResponseTime"`
	WritePhysicalIOps    float64 `json:"writePhysicalIOps"`
	WriteResponseTime    float64 `json:"writeResponseTime"`
}

type DriveStatistics struct {
	ID                string  `json:"diskId"`
	IdleTime          float64 `json:"idleTime"`
	OtherOPs          float64 `json:"otherOps"`
	OtherTimeTotal    float64 `json:"otherTimeTotal"`
	ReadBytes         float64 `json:"readBytes"`
	ReadOPs           float64 `json:"readOps"`
	ReadTimeTotal     float64 `json:"readTimeTotal"`
	RecoveredErrors   float64 `json:"recoveredErrors"`
	RetriedIOs        float64 `json:"retriedIos"`
	Timeouts          float64 `json:"timeouts"`
	UnrecoveredErrors float64 `json:"unrecoveredErrors"`
	WriteBytes        float64 `json:"writeBytes"`
	WriteOPs          float64 `json:"writeOps"`
	WriteTimeTotal    float64 `json:"writeTimeTotal"`
	QueueDepthTotal   float64 `json:"queueDepthTotal"`
	RandomIOsTotal    float64 `json:"randomIosTotal"`
	RandomBytesTotal  float64 `json:"randomBytesTotal"`
}

type DriveStatisticsCollector struct {
	AverageReadOpSize    *prometheus.Desc
	AverageWriteOpSize   *prometheus.Desc
	CombinedResponseTime *prometheus.Desc
	ReadPhysicalIOps     *prometheus.Desc
	ReadResponseTime     *prometheus.Desc
	WritePhysicalIOps    *prometheus.Desc
	WriteResponseTime    *prometheus.Desc
	IdleTime             *prometheus.Desc
	OtherOPs             *prometheus.Desc
	OtherTimeTotal       *prometheus.Desc
	ReadBytes            *prometheus.Desc
	ReadOPs              *prometheus.Desc
	ReadTimeTotal        *prometheus.Desc
	RecoveredErrors      *prometheus.Desc
	RetriedIOs           *prometheus.Desc
	Timeouts             *prometheus.Desc
	UnrecoveredErrors    *prometheus.Desc
	WriteBytes           *prometheus.Desc
	WriteOPs             *prometheus.Desc
	WriteTimeTotal       *prometheus.Desc
	QueueDepthTotal      *prometheus.Desc
	RandomIOsTotal       *prometheus.Desc
	RandomBytesTotal     *prometheus.Desc
	target               config.Target
	logger               log.Logger
}

func init() {
	registerCollector("drive-statistics", false, NewDriveStatisticsExporter)
}

func NewDriveStatisticsExporter(target config.Target, logger log.Logger) Collector {
	return &DriveStatisticsCollector{
		AverageReadOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "average_read_op_size_bytes"),
			"Drive statistic averageReadOpSize", []string{"tray", "slot"}, nil),
		AverageWriteOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "average_write_op_size_bytes"),
			"Drive statistic averageWriteOpSize", []string{"tray", "slot"}, nil),
		CombinedResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "combined_response_time_seconds"),
			"Drive statistic combinedResponseTime", []string{"tray", "slot"}, nil),
		ReadPhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_physical_iops"),
			"Drive statistic readPhysicalIOps", []string{"tray", "slot"}, nil),
		ReadResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_response_time_seconds"),
			"Drive statistic readResponseTime", []string{"tray", "slot"}, nil),
		WritePhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_physical_iops"),
			"Drive statistic writePhysicalIOps", []string{"tray", "slot"}, nil),
		WriteResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_response_time_seconds"),
			"Drive statistic writeResponseTime", []string{"tray", "slot"}, nil),
		IdleTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "idle_time_seconds_total"),
			"Drive statistic idleTime", []string{"tray", "slot"}, nil),
		OtherOPs: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "other_ops_total"),
			"Drive statistic otherOps", []string{"tray", "slot"}, nil),
		OtherTimeTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "other_time_seconds_total"),
			"Drive statistic otherTimeTotal", []string{"tray", "slot"}, nil),
		ReadBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_bytes_total"),
			"Drive statistic readBytes", []string{"tray", "slot"}, nil),
		ReadOPs: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_ops_total"),
			"Drive statistic readOps", []string{"tray", "slot"}, nil),
		ReadTimeTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "read_time_seconds_total"),
			"Drive statistic readTimeTotal", []string{"tray", "slot"}, nil),
		RecoveredErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "recovered_errors_total"),
			"Drive statistic recoveredErrors", []string{"tray", "slot"}, nil),
		RetriedIOs: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "retried_ios_total"),
			"Drive statistic retriedIos", []string{"tray", "slot"}, nil),
		Timeouts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "timeouts_total"),
			"Drive statistic timeouts", []string{"tray", "slot"}, nil),
		UnrecoveredErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "unrecovered_errors_total"),
			"Drive statistic unrecoveredErrors", []string{"tray", "slot"}, nil),
		WriteBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_bytes_total"),
			"Drive statistic writeBytes", []string{"tray", "slot"}, nil),
		WriteOPs: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_ops_total"),
			"Drive statistic writeOPs", []string{"tray", "slot"}, nil),
		WriteTimeTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "write_time_seconds_total"),
			"Drive statistic writeTimeTotal", []string{"tray", "slot"}, nil),
		QueueDepthTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "queue_depth_total"),
			"Drive statistic queueDepthTotal", []string{"tray", "slot"}, nil),
		RandomIOsTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "random_ios_total"),
			"Drive statistic randomIosTotal", []string{"tray", "slot"}, nil),
		RandomBytesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "random_bytes_total"),
			"Drive statistic randomBytesTotal", []string{"tray", "slot"}, nil),
		target: target,
		logger: logger,
	}
}

func (c *DriveStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.AverageReadOpSize
	ch <- c.AverageWriteOpSize
	ch <- c.CombinedResponseTime
	ch <- c.ReadPhysicalIOps
	ch <- c.ReadResponseTime
	ch <- c.WritePhysicalIOps
	ch <- c.WriteResponseTime
	ch <- c.IdleTime
	ch <- c.OtherOPs
	ch <- c.OtherTimeTotal
	ch <- c.ReadBytes
	ch <- c.ReadOPs
	ch <- c.ReadTimeTotal
	ch <- c.RecoveredErrors
	ch <- c.RetriedIOs
	ch <- c.Timeouts
	ch <- c.UnrecoveredErrors
	ch <- c.WriteBytes
	ch <- c.WriteOPs
	ch <- c.WriteTimeTotal
	ch <- c.QueueDepthTotal
	ch <- c.RandomIOsTotal
	ch <- c.RandomBytesTotal
}

func (c *DriveStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting drive-statistics metrics")
	collectTime := time.Now()
	var errorMetric int
	inventory, analysedDriveStatistics, driveStatistics, err := c.collect()
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

	for _, s := range analysedDriveStatistics {
		drive, ok := drives[s.ID]
		if !ok {
			drive = Drive{Slot: s.ID}
		}
		ch <- prometheus.MustNewConstMetric(c.AverageReadOpSize, prometheus.GaugeValue, s.AverageReadOpSize, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.AverageWriteOpSize, prometheus.GaugeValue, s.AverageWriteOpSize, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.CombinedResponseTime, prometheus.GaugeValue, s.CombinedResponseTime, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadPhysicalIOps, prometheus.GaugeValue, s.ReadPhysicalIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadResponseTime, prometheus.GaugeValue, s.ReadResponseTime, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WritePhysicalIOps, prometheus.GaugeValue, s.WritePhysicalIOps, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteResponseTime, prometheus.GaugeValue, s.WriteResponseTime, drive.TrayID, drive.Slot)
	}
	for _, s := range driveStatistics {
		drive, ok := drives[s.ID]
		if !ok {
			drive = Drive{Slot: s.ID}
		}
		ch <- prometheus.MustNewConstMetric(c.IdleTime, prometheus.CounterValue, s.IdleTime, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.OtherOPs, prometheus.CounterValue, s.OtherOPs, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.OtherTimeTotal, prometheus.CounterValue, s.OtherTimeTotal, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadBytes, prometheus.CounterValue, s.ReadBytes, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadOPs, prometheus.CounterValue, s.ReadOPs, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.ReadTimeTotal, prometheus.CounterValue, s.ReadTimeTotal, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.RecoveredErrors, prometheus.CounterValue, s.RecoveredErrors, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.RetriedIOs, prometheus.CounterValue, s.RetriedIOs, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.Timeouts, prometheus.CounterValue, s.Timeouts, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.UnrecoveredErrors, prometheus.CounterValue, s.UnrecoveredErrors, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteBytes, prometheus.CounterValue, s.WriteBytes, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteOPs, prometheus.CounterValue, s.WriteOPs, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.WriteTimeTotal, prometheus.CounterValue, s.WriteTimeTotal, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.QueueDepthTotal, prometheus.CounterValue, s.QueueDepthTotal, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.RandomIOsTotal, prometheus.CounterValue, s.RandomIOsTotal, drive.TrayID, drive.Slot)
		ch <- prometheus.MustNewConstMetric(c.RandomBytesTotal, prometheus.CounterValue, s.RandomBytesTotal, drive.TrayID, drive.Slot)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "drive-statistics")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "drive-statistics")
}

func (c *DriveStatisticsCollector) collect() (DrivesInventory, []AnalysedDriveStatistics, []DriveStatistics, error) {
	var inventory DrivesInventory
	var analysedDriveStatistics []AnalysedDriveStatistics
	var driveStatistics []DriveStatistics
	var inventoryBody, analyzedStatisticsBody, driveStatisticsBody []byte
	var inventoryErr, analyzedStatisticsErr, driveStatisticsErr error
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		inventoryBody, inventoryErr = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/hardware-inventory", c.target.Name), c.logger)
	}()
	go func() {
		defer wg.Done()
		analyzedStatisticsBody, analyzedStatisticsErr = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/analysed-drive-statistics", c.target.Name), c.logger)
	}()
	go func() {
		defer wg.Done()
		driveStatisticsBody, driveStatisticsErr = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/drive-statistics", c.target.Name), c.logger)
	}()
	wg.Wait()
	if inventoryErr != nil {
		return inventory, nil, nil, inventoryErr
	}
	if analyzedStatisticsErr != nil {
		return inventory, nil, nil, analyzedStatisticsErr
	}
	if driveStatisticsErr != nil {
		return inventory, nil, nil, driveStatisticsErr
	}
	err := json.Unmarshal(inventoryBody, &inventory)
	if err != nil {
		return inventory, nil, nil, err
	}
	if len(inventory.Drives) == 0 {
		return inventory, nil, nil, fmt.Errorf("No drives returned")
	}
	err = json.Unmarshal(analyzedStatisticsBody, &analysedDriveStatistics)
	if err != nil {
		return inventory, nil, nil, err
	}
	err = json.Unmarshal(driveStatisticsBody, &driveStatistics)
	if err != nil {
		return inventory, nil, nil, err
	}
	for i := range analysedDriveStatistics {
		s := &analysedDriveStatistics[i]
		// Convert milliseconds to seconds
		s.CombinedResponseTime = s.CombinedResponseTime / 1000
		s.ReadResponseTime = s.ReadResponseTime / 1000
		s.WriteResponseTime = s.WriteResponseTime / 1000
	}
	for i := range driveStatistics {
		s := &driveStatistics[i]
		// Convert microseconds to seconds
		s.IdleTime = s.IdleTime / 1000000
		s.OtherTimeTotal = s.OtherTimeTotal / 1000000
		s.ReadTimeTotal = s.ReadTimeTotal / 1000000
		s.WriteTimeTotal = s.WriteTimeTotal / 1000000
	}
	return inventory, analysedDriveStatistics, driveStatistics, nil
}
