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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/treydock/eseries_exporter/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "eseries"
)

var (
	exporterUseCache = kingpin.Flag("exporter.use-cache", "Use cached metrics if commands timeout or produce errors").Default("false").Bool()
	collectorState   = make(map[string]bool)
	factories        = make(map[string]func(target config.Target, logger log.Logger, useCache bool) Collector)
	collectDuration  = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil)
	collectError = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collect_error"),
		"Indicates if error has occurred during collection",
		[]string{"collector"}, nil)
)

type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
}

type EseriesCollector struct {
	Collectors map[string]Collector
}

func registerCollector(collector string, isDefaultEnabled bool, factory func(target config.Target, logger log.Logger, useCache bool) Collector) {
	collectorState[collector] = isDefaultEnabled
	factories[collector] = factory
}

func NewCollector(target config.Target, logger log.Logger) *EseriesCollector {
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		enable := false
		if target.Collectors == nil && enabled {
			enable = true
		} else if sliceContains(target.Collectors, key) {
			enable = true
		}
		var collector Collector
		if enable {
			collector = factories[key](target, log.With(logger, "collector", key, "target", target.Name), *exporterUseCache)
			collectors[key] = collector
		}
	}
	return &EseriesCollector{Collectors: collectors}
}

func sliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func statusToFloat64(data string) float64 {
	if bytes.Equal([]byte(data), []byte("optimal")) {
		return 1
	} else {
		return 0
	}
}

func getRequest(target config.Target, path string, logger log.Logger) ([]byte, error) {
	rel := &url.URL{Path: path}
	u := target.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(target.User, target.Password)

	level.Debug(logger).Log("msg", "Performing GET request", "url", u.String())
	resp, err := target.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		level.Error(logger).Log("msg", "Response error", "code", resp.StatusCode, "body", body)
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}
