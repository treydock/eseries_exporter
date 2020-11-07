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
	"math"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

type SystemStatistics struct {
	ID                           string  `json:"storageSystemId"`
	AverageReadOpSize            float64 `json:"averageReadOpSize"`
	AverageWriteOpSize           float64 `json:"averageWriteOpSize"`
	CacheHitBytesPercent         float64 `json:"cacheHitBytesPercent"`
	CombinedHitResponseTime      float64 `json:"combinedHitResponseTime"`
	CombinedIOps                 float64 `json:"combinedIOps"`
	CombinedResponseTime         float64 `json:"combinedResponseTime"`
	CombinedThroughput           float64 `json:"combinedThroughput"`
	CpuAvgUtilization            float64 `json:"cpuAvgUtilization"`
	DdpBytesPercent              float64 `json:"ddpBytesPercent"`
	FullStripeWritesBytesPercent float64 `json:"fullStripeWritesBytesPercent"`
	MaxCpuUtilization            float64 `json:"maxCpuUtilization"`
	RandomIosPercent             float64 `json:"randomIosPercent"`
	ReadHitResponseTime          float64 `json:"readHitResponseTime"`
	ReadIOps                     float64 `json:"readIOps"`
	ReadPhysicalIOps             float64 `json:"readPhysicalIOps"`
	ReadResponseTime             float64 `json:"readResponseTime"`
	ReadThroughput               float64 `json:"readThroughput"`
	WriteHitResponseTime         float64 `json:"writeHitResponseTime"`
	WriteIOps                    float64 `json:"writeIOps"`
	WritePhysicalIOps            float64 `json:"writePhysicalIOps"`
	WriteResponseTime            float64 `json:"writeResponseTime"`
	WriteThroughput              float64 `json:"writeThroughput"`
}

type SystemStatisticsCollector struct {
	AverageReadOpSize            *prometheus.Desc
	AverageWriteOpSize           *prometheus.Desc
	CacheHitBytesPercent         *prometheus.Desc
	CombinedHitResponseTime      *prometheus.Desc
	CombinedIOps                 *prometheus.Desc
	CombinedResponseTime         *prometheus.Desc
	CombinedThroughput           *prometheus.Desc
	CpuAvgUtilization            *prometheus.Desc
	DdpBytesPercent              *prometheus.Desc
	FullStripeWritesBytesPercent *prometheus.Desc
	MaxCpuUtilization            *prometheus.Desc
	RandomIosPercent             *prometheus.Desc
	ReadHitResponseTime          *prometheus.Desc
	ReadIOps                     *prometheus.Desc
	ReadPhysicalIOps             *prometheus.Desc
	ReadResponseTime             *prometheus.Desc
	ReadThroughput               *prometheus.Desc
	WriteHitResponseTime         *prometheus.Desc
	WriteIOps                    *prometheus.Desc
	WritePhysicalIOps            *prometheus.Desc
	WriteResponseTime            *prometheus.Desc
	WriteThroughput              *prometheus.Desc
	target                       config.Target
	logger                       log.Logger
}

func init() {
	registerCollector("system-statistics", true, NewSystemStatisticsExporter)
}

func NewSystemStatisticsExporter(target config.Target, logger log.Logger) Collector {
	return &SystemStatisticsCollector{
		AverageReadOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "average_read_op_size_bytes"),
			"System statistic averageReadOpSize", []string{"id"}, nil),
		AverageWriteOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "average_write_op_size_bytes"),
			"System statistic averageWriteOpSize", []string{"id"}, nil),
		CacheHitBytesPercent: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "cache_hit_bytes_percent"),
			"System statistic CacheHitBytesPercent", []string{"id"}, nil),
		CombinedHitResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "combined_hit_response_time_seconds"),
			"System statistic CombinedHitResponseTime", []string{"id"}, nil),
		CombinedIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "combined_iops"),
			"System statistic combinedIOps", []string{"id"}, nil),
		CombinedResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "combined_response_time_seconds"),
			"System statistic combinedResponseTime", []string{"id"}, nil),
		CombinedThroughput: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "combined_throughput_bytes_per_second"),
			"System statistic combinedThroughput", []string{"id"}, nil),
		CpuAvgUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "cpu_avg_utilization"),
			"System statistic CpuAvgUtilization", []string{"id"}, nil),
		DdpBytesPercent: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "ddp_bytes_percent"),
			"System statistic DdpBytesPercent", []string{"id"}, nil),
		FullStripeWritesBytesPercent: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "full_stripe_writes_bytes_percent"),
			"System statistic FullStripeWritesBytesPercent", []string{"id"}, nil),
		MaxCpuUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "max_cpu_utilization"),
			"System statistic MaxCpuUtilization", []string{"id"}, nil),
		RandomIosPercent: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "random_ios_percent"),
			"System statistic RandomIosPercent", []string{"id"}, nil),
		ReadHitResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_hit_response_time_seconds"),
			"System statistic ReadHitResponseTime", []string{"id"}, nil),
		ReadIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_iops"),
			"System statistic readIOps", []string{"id"}, nil),
		ReadPhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_physical_iops"),
			"System statistic readPhysicalIOps", []string{"id"}, nil),
		ReadResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_response_time_seconds"),
			"System statistic readResponseTime", []string{"id"}, nil),
		ReadThroughput: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "read_throughput_bytes_per_second"),
			"System statistic combinedThroughput", []string{"id"}, nil),
		WriteHitResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_hit_response_time_seconds"),
			"System statistic WriteHitResponseTime", []string{"id"}, nil),
		WriteIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_iops"),
			"System statistic writeIOps", []string{"id"}, nil),
		WritePhysicalIOps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_physical_iops"),
			"System statistic writePhysicalIOps", []string{"id"}, nil),
		WriteResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_response_time_seconds"),
			"System statistic writeResponseTime", []string{"id"}, nil),
		WriteThroughput: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "write_throughput_bytes_per_second"),
			"System statistic combinedThroughput", []string{"id"}, nil),
		target: target,
		logger: logger,
	}
}

func (c *SystemStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.AverageReadOpSize
	ch <- c.AverageWriteOpSize
	ch <- c.CacheHitBytesPercent
	ch <- c.CombinedHitResponseTime
	ch <- c.CombinedIOps
	ch <- c.CombinedResponseTime
	ch <- c.CombinedThroughput
	ch <- c.CpuAvgUtilization
	ch <- c.DdpBytesPercent
	ch <- c.FullStripeWritesBytesPercent
	ch <- c.MaxCpuUtilization
	ch <- c.RandomIosPercent
	ch <- c.ReadHitResponseTime
	ch <- c.ReadIOps
	ch <- c.ReadPhysicalIOps
	ch <- c.ReadResponseTime
	ch <- c.ReadThroughput
	ch <- c.WriteHitResponseTime
	ch <- c.WriteIOps
	ch <- c.WritePhysicalIOps
	ch <- c.WriteResponseTime
	ch <- c.WriteThroughput
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
		ch <- prometheus.MustNewConstMetric(c.AverageReadOpSize, prometheus.GaugeValue, statistics.AverageReadOpSize, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.AverageWriteOpSize, prometheus.GaugeValue, statistics.AverageWriteOpSize, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.CacheHitBytesPercent, prometheus.GaugeValue, statistics.CacheHitBytesPercent, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.CombinedHitResponseTime, prometheus.GaugeValue, statistics.CombinedHitResponseTime, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.CombinedIOps, prometheus.GaugeValue, statistics.CombinedIOps, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.CombinedResponseTime, prometheus.GaugeValue, statistics.CombinedResponseTime, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.CombinedThroughput, prometheus.GaugeValue, statistics.CombinedThroughput, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.CpuAvgUtilization, prometheus.GaugeValue, statistics.CpuAvgUtilization, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.DdpBytesPercent, prometheus.GaugeValue, statistics.DdpBytesPercent, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.FullStripeWritesBytesPercent, prometheus.GaugeValue, statistics.FullStripeWritesBytesPercent, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.MaxCpuUtilization, prometheus.GaugeValue, statistics.MaxCpuUtilization, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.RandomIosPercent, prometheus.GaugeValue, statistics.RandomIosPercent, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.ReadHitResponseTime, prometheus.GaugeValue, statistics.ReadHitResponseTime, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.ReadIOps, prometheus.GaugeValue, statistics.ReadIOps, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.ReadPhysicalIOps, prometheus.GaugeValue, statistics.ReadPhysicalIOps, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.ReadResponseTime, prometheus.GaugeValue, statistics.ReadResponseTime, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.ReadThroughput, prometheus.GaugeValue, statistics.ReadThroughput, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.WriteHitResponseTime, prometheus.GaugeValue, statistics.WriteHitResponseTime, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.WriteIOps, prometheus.GaugeValue, statistics.WriteIOps, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.WritePhysicalIOps, prometheus.GaugeValue, statistics.WritePhysicalIOps, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.WriteResponseTime, prometheus.GaugeValue, statistics.WriteResponseTime, statistics.ID)
		ch <- prometheus.MustNewConstMetric(c.WriteThroughput, prometheus.GaugeValue, statistics.WriteThroughput, statistics.ID)
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
	// Convert MB/s to bytes/s
	statistics.CombinedThroughput = statistics.CombinedThroughput * math.Pow(1024, 2)
	statistics.ReadThroughput = statistics.ReadThroughput * math.Pow(1024, 2)
	statistics.WriteThroughput = statistics.WriteThroughput * math.Pow(1024, 2)
	return statistics, nil
}
