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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

type SystemStatistics struct {
	AverageReadOpSize       float64 `json:"averageReadOpSize"`
	AverageWriteOpSize      float64 `json:"averageWriteOpSize"`
	CombinedHitResponseTime float64 `json:"combinedHitResponseTime"`
	CombinedResponseTime    float64 `json:"combinedResponseTime"`
	CpuAvgUtilization       float64 `json:"cpuAvgUtilization"`
	MaxCpuUtilization       float64 `json:"maxCpuUtilization"`
	ReadHitResponseTime     float64 `json:"readHitResponseTime"`
	ReadPhysicalIOps        float64 `json:"readPhysicalIOps"`
	ReadResponseTime        float64 `json:"readResponseTime"`
	WriteHitResponseTime    float64 `json:"writeHitResponseTime"`
	WritePhysicalIOps       float64 `json:"writePhysicalIOps"`
	WriteResponseTime       float64 `json:"writeResponseTime"`
}

type SystemStatisticsCollector struct {
	AverageReadOpSize       *prometheus.Desc
	AverageWriteOpSize      *prometheus.Desc
	CombinedHitResponseTime *prometheus.Desc
	CombinedResponseTime    *prometheus.Desc
	CpuAvgUtilization       *prometheus.Desc
	MaxCpuUtilization       *prometheus.Desc
	ReadHitResponseTime     *prometheus.Desc
	ReadPhysicalIOps        *prometheus.Desc
	ReadResponseTime        *prometheus.Desc
	WriteHitResponseTime    *prometheus.Desc
	WritePhysicalIOps       *prometheus.Desc
	WriteResponseTime       *prometheus.Desc
	target                  config.Target
	logger                  log.Logger
}

func init() {
	registerCollector("system-statistics", true, NewSystemStatisticsExporter)
}

func NewSystemStatisticsExporter(target config.Target, logger log.Logger) Collector {
	return &SystemStatisticsCollector{
		AverageReadOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "average_read_op_size_bytes"),
			"System statistic averageReadOpSize", nil, nil),
		AverageWriteOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "average_write_op_size_bytes"),
			"System statistic averageWriteOpSize", nil, nil),
		CombinedHitResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "combined_hit_response_time_seconds"),
			"System statistic CombinedHitResponseTime", nil, nil),
		CombinedResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "combined_response_time_seconds"),
			"System statistic combinedResponseTime", nil, nil),
		CpuAvgUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "cpu_average_utilization"),
			"System statistic CpuAvgUtilization (0.0-1.0 ratio of CPU percent utilization)", nil, nil),
		MaxCpuUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "cpu_max_utilization"),
			"System statistic MaxCpuUtilization (0.0-1.0 ratio of CPU percent utilization)", nil, nil),
		ReadHitResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_hit_response_time_seconds"),
			"System statistic ReadHitResponseTime", nil, nil),
		ReadPhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_physical_iops"),
			"System statistic readPhysicalIOps", nil, nil),
		ReadResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_response_time_seconds"),
			"System statistic readResponseTime", nil, nil),
		WriteHitResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_hit_response_time_seconds"),
			"System statistic WriteHitResponseTime", nil, nil),
		WritePhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_physical_iops"),
			"System statistic writePhysicalIOps", nil, nil),
		WriteResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_response_time_seconds"),
			"System statistic writeResponseTime", nil, nil),
		target: target,
		logger: logger,
	}
}

func (c *SystemStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.AverageReadOpSize
	ch <- c.AverageWriteOpSize
	ch <- c.CombinedHitResponseTime
	ch <- c.CombinedResponseTime
	ch <- c.CpuAvgUtilization
	ch <- c.MaxCpuUtilization
	ch <- c.ReadHitResponseTime
	ch <- c.ReadPhysicalIOps
	ch <- c.ReadResponseTime
	ch <- c.WriteHitResponseTime
	ch <- c.WritePhysicalIOps
	ch <- c.WriteResponseTime
}

func (c *SystemStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting system-statistics metrics")
	collectTime := time.Now()
	var errorMetric int
	statistics, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	if err == nil {
		ch <- prometheus.MustNewConstMetric(c.AverageReadOpSize, prometheus.GaugeValue, statistics.AverageReadOpSize)
		ch <- prometheus.MustNewConstMetric(c.AverageWriteOpSize, prometheus.GaugeValue, statistics.AverageWriteOpSize)
		ch <- prometheus.MustNewConstMetric(c.CombinedHitResponseTime, prometheus.GaugeValue, statistics.CombinedHitResponseTime)
		ch <- prometheus.MustNewConstMetric(c.CombinedResponseTime, prometheus.GaugeValue, statistics.CombinedResponseTime)
		ch <- prometheus.MustNewConstMetric(c.CpuAvgUtilization, prometheus.GaugeValue, statistics.CpuAvgUtilization)
		ch <- prometheus.MustNewConstMetric(c.MaxCpuUtilization, prometheus.GaugeValue, statistics.MaxCpuUtilization)
		ch <- prometheus.MustNewConstMetric(c.ReadHitResponseTime, prometheus.GaugeValue, statistics.ReadHitResponseTime)
		ch <- prometheus.MustNewConstMetric(c.ReadPhysicalIOps, prometheus.GaugeValue, statistics.ReadPhysicalIOps)
		ch <- prometheus.MustNewConstMetric(c.ReadResponseTime, prometheus.GaugeValue, statistics.ReadResponseTime)
		ch <- prometheus.MustNewConstMetric(c.WriteHitResponseTime, prometheus.GaugeValue, statistics.WriteHitResponseTime)
		ch <- prometheus.MustNewConstMetric(c.WritePhysicalIOps, prometheus.GaugeValue, statistics.WritePhysicalIOps)
		ch <- prometheus.MustNewConstMetric(c.WriteResponseTime, prometheus.GaugeValue, statistics.WriteResponseTime)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "system-statistics")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "system-statistics")
}

func (c *SystemStatisticsCollector) collect() (SystemStatistics, error) {
	var statistics SystemStatistics
	statisticsBody, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/analysed-system-statistics", c.target.Name), c.logger)
	if err != nil {
		return statistics, err
	}
	err = json.Unmarshal(statisticsBody, &statistics)
	if err != nil {
		return statistics, err
	}
	// Convert milliseconds to seconds
	statistics.CombinedHitResponseTime = statistics.CombinedHitResponseTime * 0.001
	statistics.CombinedResponseTime = statistics.CombinedResponseTime * 0.001
	statistics.ReadHitResponseTime = statistics.ReadHitResponseTime * 0.001
	statistics.ReadResponseTime = statistics.ReadResponseTime * 0.001
	statistics.WriteHitResponseTime = statistics.WriteHitResponseTime * 0.001
	statistics.WriteResponseTime = statistics.WriteResponseTime * 0.001
	// Convert from percent to ratio
	statistics.MaxCpuUtilization = statistics.MaxCpuUtilization / 100
	statistics.CpuAvgUtilization = statistics.CpuAvgUtilization / 100
	return statistics, nil
}
