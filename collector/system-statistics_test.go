// MIT License
//
// Copyright (c) 2020 Ohio Supercomputer Center
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package collector

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/treydock/eseries_exporter/config"
)

func TestSystemStatisticsCollector(t *testing.T) {
	fixtureData, err := os.ReadFile("testdata/system-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="system-statistics"} 0
	# HELP eseries_system_average_read_op_size_bytes System statistic averageReadOpSize
	# TYPE eseries_system_average_read_op_size_bytes gauge
	eseries_system_average_read_op_size_bytes 17357.11013434037
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test",
		User:       "test",
		Password:   "test",
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewSystemStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 14 {
		t.Errorf("Unexpected collection count %d, expected 14", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		/*
			"eseries_system_average_read_op_size", "eseries_system_average_write_op_size",
			"eseries_system_combined_iops", "eseries_system_combined_response_time", "eseries_system_combined_throughput",
			"eseries_system_read_iops", "eseries_system_read_ops", "eseries_system_read_physical_iops",
			"eseries_system_read_response_time", "eseries_system_read_throughput",
			"eseries_system_write_iops", "eseries_system_write_ops", "eseries_system_write_physical_iops",
			"eseries_system_write_response_time", "eseries_system_write_throughput",
		*/
		"eseries_system_average_read_op_size_bytes",
		"eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSystemStatisticsCollectorError(t *testing.T) {
	expected := `
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="system-statistics"} 1
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "error", http.StatusNotFound)
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test",
		User:       "test",
		Password:   "test",
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewSystemStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_system_average_read_op_size_bytes", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
