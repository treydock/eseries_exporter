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

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/treydock/eseries_exporter/collector"
	"github.com/treydock/eseries_exporter/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile    = kingpin.Flag("config.file", "Path to exporter config file").Default("eseries_exporter.yaml").String()
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9313").String()
)

func metricsHandler(c *config.Config, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()

		t := r.URL.Query().Get("target")
		if t == "" {
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			return
		}
		m := r.URL.Query().Get("module")
		if m == "" {
			m = "default"
		}
		module, ok := c.Modules[m]
		if !ok {
			http.Error(w, fmt.Sprintf("Unknown module %s", t), http.StatusNotFound)
			return
		}
		target := config.Target{
			Name:       t,
			User:       module.User,
			Password:   module.Password,
			ProxyURL:   module.ProxyURL,
			Collectors: module.Collectors,
		}
		if url, err := url.Parse(module.ProxyURL); err != nil {
			http.Error(w, fmt.Sprintf("Unable to parse ProxyURL %s", module.ProxyURL), http.StatusNotFound)
			return
		} else {
			target.BaseURL = url
		}
		httpClient := &http.Client{
			Timeout: time.Duration(module.Timeout) * time.Second,
		}
		target.HttpClient = httpClient
		eseriesCollector := collector.NewCollector(target, logger)
		for key, collector := range eseriesCollector.Collectors {
			level.Debug(logger).Log("msg", fmt.Sprintf("Enabled collector %s", key))
			registry.MustRegister(collector)
		}

		gatherers := prometheus.Gatherers{registry}

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	metricsEndpoint := "/eseries"
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("eseries_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)
	level.Info(logger).Log("msg", "Starting eseries_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	level.Info(logger).Log("msg", "Starting Server", "address", *listenAddress)

	sc := &config.SafeConfig{}

	if err := sc.ReloadConfig(*configFile); err != nil {
		level.Error(logger).Log("msg", "Error loading config", "err", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		w.Write([]byte(`<html>
             <head><title>E-Series Exporter</title></head>
             <body>
             <h1>E-Series Exporter</h1>
             <p><a href='` + metricsEndpoint + `'>E-Series Metrics</a></p>
             <p><a href='/metrics'>Exporter Metrics</a></p>
             </body>
             </html>`))
	})
	http.Handle(metricsEndpoint, metricsHandler(sc.C, logger))
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
