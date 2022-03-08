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
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

type AnalysedControllerStatistics struct {
	ID                   string `json:"controllerId"`
	Label                string
	AverageReadOpSize    float64 `json:"averageReadOpSize"`
	AverageWriteOpSize   float64 `json:"averageWriteOpSize"`
	CombinedResponseTime float64 `json:"combinedResponseTime"`
	ReadResponseTime     float64 `json:"readResponseTime"`
	WriteResponseTime    float64 `json:"writeResponseTime"`
	MaxCpuUtilization    float64 `json:"maxCpuUtilization"`
	CpuAvgUtilization    float64 `json:"cpuAvgUtilization"`
}

type ControllerStatistics struct {
	ID                              string `json:"controllerId"`
	Label                           string
	TotalIopsServiced               float64 `json:"totalIopsServiced"`
	TotalBytesServiced              float64 `json:"totalBytesServiced"`
	CacheHitsIopsTotal              float64 `json:"cacheHitsIopsTotal"`
	CacheHitsBytesTotal             float64 `json:"cacheHitsBytesTotal"`
	RandomIosTotal                  float64 `json:"randomIosTotal"`
	RandomBytesTotal                float64 `json:"randomBytesTotal"`
	ReadIopsTotal                   float64 `json:"readIopsTotal"`
	ReadBytesTotal                  float64 `json:"readBytesTotal"`
	WriteIopsTotal                  float64 `json:"writeIopsTotal"`
	WriteBytesTotal                 float64 `json:"writeBytesTotal"`
	MirrorIopsTotal                 float64 `json:"mirrorIopsTotal"`
	MirrorBytesTotal                float64 `json:"mirrorBytesTotal"`
	FullStripeWritesBytes           float64 `json:"fullStripeWritesBytes"`
	Raid0BytesTransferred           float64 `json:"raid0BytesTransferred"`
	Raid1BytesTransferred           float64 `json:"raid1BytesTransferred"`
	Raid5BytesTransferred           float64 `json:"raid5BytesTransferred"`
	Raid6BytesTransferred           float64 `json:"raid6BytesTransferred"`
	DdpBytesTransferred             float64 `json:"ddpBytesTransferred"`
	MaxPossibleBpsUnderCurrentLoad  float64 `json:"maxPossibleBpsUnderCurrentLoad"`
	MaxPossibleIopsUnderCurrentLoad float64 `json:"maxPossibleIopsUnderCurrentLoad"`
}

type ControllersInventory struct {
	Controllers []Controller `json:"controllers"`
}

type Controller struct {
	ID               string                     `json:"id"`
	PhysicalLocation ControllerPhysicalLocation `json:"physicalLocation"`
	Label            string
}

type ControllerPhysicalLocation struct {
	Label string `json:"label"`
}

type ControllerStatisticsCollector struct {
	AverageReadOpSize               *prometheus.Desc
	AverageWriteOpSize              *prometheus.Desc
	CombinedResponseTime            *prometheus.Desc
	ReadResponseTime                *prometheus.Desc
	WriteResponseTime               *prometheus.Desc
	MaxCpuUtilization               *prometheus.Desc
	CpuAvgUtilization               *prometheus.Desc
	TotalIopsServiced               *prometheus.Desc
	TotalBytesServiced              *prometheus.Desc
	CacheHitsIopsTotal              *prometheus.Desc
	CacheHitsBytesTotal             *prometheus.Desc
	RandomIosTotal                  *prometheus.Desc
	RandomBytesTotal                *prometheus.Desc
	ReadIopsTotal                   *prometheus.Desc
	ReadBytesTotal                  *prometheus.Desc
	WriteIopsTotal                  *prometheus.Desc
	WriteBytesTotal                 *prometheus.Desc
	MirrorIopsTotal                 *prometheus.Desc
	MirrorBytesTotal                *prometheus.Desc
	FullStripeWritesBytes           *prometheus.Desc
	Raid0BytesTransferred           *prometheus.Desc
	Raid1BytesTransferred           *prometheus.Desc
	Raid5BytesTransferred           *prometheus.Desc
	Raid6BytesTransferred           *prometheus.Desc
	DdpBytesTransferred             *prometheus.Desc
	MaxPossibleBpsUnderCurrentLoad  *prometheus.Desc
	MaxPossibleIopsUnderCurrentLoad *prometheus.Desc
	target                          config.Target
	logger                          log.Logger
}

func init() {
	registerCollector("controller-statistics", true, NewControllerStatisticsExporter)
}

func NewControllerStatisticsExporter(target config.Target, logger log.Logger) Collector {
	labels := []string{"controller", "controller_label"}
	return &ControllerStatisticsCollector{
		AverageReadOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "average_read_op_size_bytes"),
			"Controller statistic averageReadOpSize", labels, nil),
		AverageWriteOpSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "average_write_op_size_bytes"),
			"Controller statistic averageWriteOpSize", labels, nil),
		CombinedResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "combined_response_time_seconds"),
			"Controller statistic combinedResponseTime", labels, nil),
		ReadResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "read_response_time_seconds"),
			"Controller statistic readResponseTime", labels, nil),
		WriteResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "write_response_time_seconds"),
			"Controller statistic writeResponseTime", labels, nil),
		MaxCpuUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "cpu_max_utilization_ratio"),
			"Controller statistic maxCpuUtilization (0.0-1.0 ratio of CPU percent utilization)", labels, nil),
		CpuAvgUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "cpu_average_utilization_ratio"),
			"Controller statistic cpuAvgUtilization (0.0-1.0 ratio of CPU percent utilization)", labels, nil),
		TotalIopsServiced: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "iops_total"),
			"Controller statistic totalIopsServiced", labels, nil),
		TotalBytesServiced: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "bytes_total"),
			"Controller statistic totalBytesServiced", labels, nil),
		CacheHitsIopsTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "cache_hits_iops_total"),
			"Controller statistic cacheHitsIopsTotal", labels, nil),
		CacheHitsBytesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "cache_hit_bytes_total"),
			"Controller statistic cacheHitsBytesTotal", labels, nil),
		RandomIosTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "random_ios_total"),
			"Controller statistic randomIosTotal", labels, nil),
		RandomBytesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "random_bytes_total"),
			"Controller statistic randomBytesTotal", labels, nil),
		ReadIopsTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "read_iops_total"),
			"Controller statistic readIopsTotal", labels, nil),
		ReadBytesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "read_bytes_total"),
			"Controller statistic readBytesTotal", labels, nil),
		WriteIopsTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "write_iops_total"),
			"Controller statistic writeIopsTotal", labels, nil),
		WriteBytesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "write_bytes_total"),
			"Controller statistic writeBytesTotal", labels, nil),
		MirrorIopsTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "mirror_iops_total"),
			"Controller statistic mirrorIopsTotal", labels, nil),
		MirrorBytesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "mirror_bytes_total"),
			"Controller statistic mirrorBytesTotal", labels, nil),
		FullStripeWritesBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "full_stripe_writes_bytes_total"),
			"Controller statistic fullStripeWritesBytes", labels, nil),
		Raid0BytesTransferred: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "raid0_transferred_bytes_total"),
			"Controller statistic raid0BytesTransferred", labels, nil),
		Raid1BytesTransferred: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "raid1_transferred_bytes_total"),
			"Controller statistic raid1BytesTransferred", labels, nil),
		Raid5BytesTransferred: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "raid5_transferred_bytes_total"),
			"Controller statistic raid5BytesTransferred", labels, nil),
		Raid6BytesTransferred: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "raid6_transferred_bytes_total"),
			"Controller statistic raid6BytesTransferred", labels, nil),
		DdpBytesTransferred: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "ddp_transferred_bytes_total"),
			"Controller statistic ddpBytesTransferred", labels, nil),
		MaxPossibleBpsUnderCurrentLoad: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "max_possible_throughput_bytes_per_second"),
			"Controller statistic maxPossibleBpsUnderCurrentLoad", labels, nil),
		MaxPossibleIopsUnderCurrentLoad: prometheus.NewDesc(prometheus.BuildFQName(namespace, "controller", "max_possible_iops"),
			"Controller statistic maxPossibleIopsUnderCurrentLoad", labels, nil),
		target: target,
		logger: logger,
	}
}

func (c *ControllerStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.AverageReadOpSize
	ch <- c.AverageWriteOpSize
	ch <- c.CombinedResponseTime
	ch <- c.ReadResponseTime
	ch <- c.WriteResponseTime
	ch <- c.MaxCpuUtilization
	ch <- c.CpuAvgUtilization
	ch <- c.TotalIopsServiced
	ch <- c.TotalBytesServiced
	ch <- c.CacheHitsIopsTotal
	ch <- c.CacheHitsBytesTotal
	ch <- c.RandomIosTotal
	ch <- c.RandomBytesTotal
	ch <- c.ReadIopsTotal
	ch <- c.ReadBytesTotal
	ch <- c.WriteIopsTotal
	ch <- c.WriteBytesTotal
	ch <- c.MirrorIopsTotal
	ch <- c.MirrorBytesTotal
	ch <- c.FullStripeWritesBytes
	ch <- c.Raid0BytesTransferred
	ch <- c.Raid1BytesTransferred
	ch <- c.Raid5BytesTransferred
	ch <- c.Raid6BytesTransferred
	ch <- c.DdpBytesTransferred
	ch <- c.MaxPossibleBpsUnderCurrentLoad
	ch <- c.MaxPossibleIopsUnderCurrentLoad
}

func (c *ControllerStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting controller-statistics metrics")
	collectTime := time.Now()
	var errorMetric int
	analyzedStatistics, statistics, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	for _, s := range analyzedStatistics {
		ch <- prometheus.MustNewConstMetric(c.AverageReadOpSize, prometheus.GaugeValue, s.AverageReadOpSize, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.AverageWriteOpSize, prometheus.GaugeValue, s.AverageWriteOpSize, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.CombinedResponseTime, prometheus.GaugeValue, s.CombinedResponseTime, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ReadResponseTime, prometheus.GaugeValue, s.ReadResponseTime, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.WriteResponseTime, prometheus.GaugeValue, s.WriteResponseTime, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.MaxCpuUtilization, prometheus.GaugeValue, s.MaxCpuUtilization, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.CpuAvgUtilization, prometheus.GaugeValue, s.CpuAvgUtilization, s.ID, s.Label)
	}

	for _, s := range statistics {
		ch <- prometheus.MustNewConstMetric(c.TotalIopsServiced, prometheus.CounterValue, s.TotalIopsServiced, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.TotalBytesServiced, prometheus.CounterValue, s.TotalBytesServiced, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.CacheHitsIopsTotal, prometheus.CounterValue, s.CacheHitsIopsTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.CacheHitsBytesTotal, prometheus.CounterValue, s.CacheHitsBytesTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.RandomIosTotal, prometheus.CounterValue, s.RandomIosTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.RandomBytesTotal, prometheus.CounterValue, s.RandomBytesTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ReadIopsTotal, prometheus.CounterValue, s.ReadIopsTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ReadBytesTotal, prometheus.CounterValue, s.ReadBytesTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.WriteIopsTotal, prometheus.CounterValue, s.WriteIopsTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.WriteBytesTotal, prometheus.CounterValue, s.WriteBytesTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.MirrorIopsTotal, prometheus.CounterValue, s.MirrorIopsTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.MirrorBytesTotal, prometheus.CounterValue, s.MirrorBytesTotal, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.FullStripeWritesBytes, prometheus.CounterValue, s.FullStripeWritesBytes, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.Raid0BytesTransferred, prometheus.CounterValue, s.Raid0BytesTransferred, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.Raid1BytesTransferred, prometheus.CounterValue, s.Raid1BytesTransferred, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.Raid5BytesTransferred, prometheus.CounterValue, s.Raid5BytesTransferred, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.Raid6BytesTransferred, prometheus.CounterValue, s.Raid6BytesTransferred, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.DdpBytesTransferred, prometheus.CounterValue, s.DdpBytesTransferred, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.MaxPossibleBpsUnderCurrentLoad, prometheus.CounterValue, s.MaxPossibleBpsUnderCurrentLoad, s.ID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.MaxPossibleIopsUnderCurrentLoad, prometheus.CounterValue, s.MaxPossibleIopsUnderCurrentLoad, s.ID, s.Label)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "controller-statistics")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "controller-statistics")
}

func (c *ControllerStatisticsCollector) collect() ([]AnalysedControllerStatistics, []ControllerStatistics, error) {
	var inventory ControllersInventory
	var analyzedStatistics []AnalysedControllerStatistics
	var statistics []ControllerStatistics
	var inventoryBody, analyzedStatisticsBody, statisticsBody []byte
	var inventoryErr, analyzedStatisticsErr, statisticsErr error
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		inventoryBody, inventoryErr = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/hardware-inventory", c.target.Name), c.logger)
	}()
	go func() {
		defer wg.Done()
		analyzedStatisticsBody, analyzedStatisticsErr = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/analysed-controller-statistics", c.target.Name), c.logger)
	}()
	go func() {
		defer wg.Done()
		statisticsBody, statisticsErr = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/controller-statistics", c.target.Name), c.logger)
	}()
	wg.Wait()
	if inventoryErr != nil {
		return nil, nil, inventoryErr
	}
	if analyzedStatisticsErr != nil {
		return nil, nil, analyzedStatisticsErr
	}
	if statisticsErr != nil {
		return nil, nil, statisticsErr
	}
	err := json.Unmarshal(inventoryBody, &inventory)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(analyzedStatisticsBody, &analyzedStatistics)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(statisticsBody, &statistics)
	if err != nil {
		return nil, nil, err
	}
	controllers := make(map[string]Controller)
	for _, c := range inventory.Controllers {
		c.Label = c.PhysicalLocation.Label
		controllers[c.ID] = c
	}
	for i := range analyzedStatistics {
		s := &analyzedStatistics[i]
		controller, ok := controllers[s.ID]
		if ok {
			s.Label = controller.Label
		}
		// Convert milliseconds to seconds
		s.CombinedResponseTime = s.CombinedResponseTime * 0.001
		s.ReadResponseTime = s.ReadResponseTime * 0.001
		s.WriteResponseTime = s.WriteResponseTime * 0.001
		// Convert from percent to ratio
		s.MaxCpuUtilization = s.MaxCpuUtilization / 100
		s.CpuAvgUtilization = s.CpuAvgUtilization / 100
	}
	for i := range statistics {
		s := &statistics[i]
		controller, ok := controllers[s.ID]
		if ok {
			s.Label = controller.Label
		}
	}
	return analyzedStatistics, statistics, nil
}
