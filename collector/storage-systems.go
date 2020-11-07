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

var (
	storageSystemsStatuses = []string{
		"neverContacted",
		"offline",
		"optimal",
		"needsAttn",
		"removed",
		"newDevice",
		"lockDown",
	}
)

type StorageSystem struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type StorageSystemsCollector struct {
	Status *prometheus.Desc
	target config.Target
	logger log.Logger
}

func init() {
	registerCollector("storage-systems", true, NewStorageSystemsExporter)
}

func NewStorageSystemsExporter(target config.Target, logger log.Logger) Collector {
	return &StorageSystemsCollector{
		Status: prometheus.NewDesc(prometheus.BuildFQName(namespace, "storage_system", "status"),
			"Storage System status, 1=optimal 0=all other states", []string{"status"}, nil),
		target: target,
		logger: logger,
	}
}

func (c *StorageSystemsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
}

func (c *StorageSystemsCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting storage-systems metrics")
	collectTime := time.Now()
	var errorMetric int
	metric, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	if err == nil {
		for _, status := range storageSystemsStatuses {
			var value float64
			if status == metric.Status {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.Status, prometheus.GaugeValue, value, status)
		}
		var unknown float64
		if !sliceContains(storageSystemsStatuses, metric.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.Status, prometheus.GaugeValue, unknown, "unknown")
	}
	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "storage-systems")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "storage-systems")
}

func (c *StorageSystemsCollector) collect() (StorageSystem, error) {
	var metrics StorageSystem
	body, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s", c.target.Name), c.logger)
	if err != nil {
		return metrics, err
	}
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return metrics, err
	}
	if metrics.ID == "" {
		return metrics, fmt.Errorf("Not storage systems returned")
	}
	return metrics, nil
}
