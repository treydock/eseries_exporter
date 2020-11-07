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
	"io/ioutil"
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

func TestDriveStatisticsCollector(t *testing.T) {
	fixtureData, err := ioutil.ReadFile("testdata/drive-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	inventoryData, err := ioutil.ReadFile("testdata/drives.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `
	# HELP eseries_drive_average_read_op_size_bytes Drive statistic averageReadOpSize
	# TYPE eseries_drive_average_read_op_size_bytes gauge
	eseries_drive_average_read_op_size_bytes{slot="58",tray="0"} 39620.99569760295
	eseries_drive_average_read_op_size_bytes{slot="53",tray="0"} 21312.646464646463
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="drive-statistics"} 0
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "hardware-inventory") {
			_, _ = rw.Write(inventoryData)
		} else {
			_, _ = rw.Write(fixtureData)
		}
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
	collector := NewDriveStatisticsExporter(target, logger, false)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 32 {
		t.Errorf("Unexpected collection count %d, expected 32", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		/*
			"eseries_drive_average_read_op_size", "eseries_drive_average_write_op_size",
			"eseries_drive_combined_iops", "eseries_drive_combined_response_time", "eseries_drive_combined_throughput",
			"eseries_drive_read_iops", "eseries_drive_read_ops", "eseries_drive_read_physical_iops",
			"eseries_drive_read_response_time", "eseries_drive_read_throughput",
			"eseries_drive_write_iops", "eseries_drive_write_ops", "eseries_drive_write_physical_iops",
			"eseries_drive_write_response_time", "eseries_drive_write_throughput",
		*/
		"eseries_drive_average_read_op_size_bytes",
		"eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestDriveStatisticsCollectorError(t *testing.T) {
	expected := `
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="drive-statistics"} 1
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
	collector := NewDriveStatisticsExporter(target, logger, false)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_drive_average_read_op_size_bytes", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
