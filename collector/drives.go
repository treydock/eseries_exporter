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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
)

var (
	drivesCache      = map[string][]Drive{}
	drivesCacheMutex = sync.RWMutex{}
)

type Drive struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type DrivesCollector struct {
	Status   *prometheus.Desc
	target   config.Target
	logger   log.Logger
	useCache bool
}

func init() {
	registerCollector("drives", true, NewDrivesExporter)
}

func NewDrivesExporter(target config.Target, logger log.Logger, useCache bool) Collector {
	return &DrivesCollector{
		Status: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "status"),
			"Drive status, 1=optimal 0=all other states", []string{"systemid", "id", "status"}, nil),
		target:   target,
		logger:   logger,
		useCache: useCache,
	}
}

func (c *DrivesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
}

func (c *DrivesCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting drives metrics")
	collectTime := time.Now()
	var errorMetric int
	metrics, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	for _, m := range metrics {
		ch <- prometheus.MustNewConstMetric(c.Status, prometheus.GaugeValue, statusToFloat64(m.Status), c.target.Name, m.ID, m.Status)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "drives")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "drives")
}

func (c *DrivesCollector) collect() ([]Drive, error) {
	var metrics []Drive
	body, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/drives", c.target.Name), c.logger)
	if err != nil {
		if c.useCache {
			metrics = drivesReadCache(c.target.Name)
		}
		return metrics, err
	}
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		if c.useCache {
			metrics = drivesReadCache(c.target.Name)
		}
		return metrics, err
	}
	if len(metrics) == 0 {
		if c.useCache {
			metrics = drivesReadCache(c.target.Name)
		}
		return metrics, fmt.Errorf("No drives returned")
	}
	if c.useCache {
		drivesWriteCache(c.target.Name, metrics)
	}
	return metrics, nil
}

func drivesReadCache(target string) []Drive {
	var metrics []Drive
	drivesCacheMutex.RLock()
	if cache, ok := drivesCache[target]; ok {
		metrics = cache
	}
	drivesCacheMutex.RUnlock()
	return metrics
}

func drivesWriteCache(target string, metrics []Drive) {
	drivesCacheMutex.Lock()
	drivesCache[target] = metrics
	drivesCacheMutex.Unlock()
}
