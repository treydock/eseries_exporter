[![Build Status](https://circleci.com/gh/treydock/eseries_exporter/tree/master.svg?style=shield)](https://circleci.com/gh/treydock/eseries_exporter)
[![GitHub release](https://img.shields.io/github/v/release/treydock/eseries_exporter?include_prereleases&sort=semver)](https://github.com/treydock/eseries_exporter/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/treydock/eseries_exporter/total)
[![codecov](https://codecov.io/gh/treydock/eseries_exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/treydock/eseries_exporter)

# NetApp E-Series Prometheus exporter

The E-Series exporter collects metrics from NetApp E-Series via the SANtricity Web Services Proxy.

This exporter is intended to query multiple E-Series controllers from an external host.

The `/eseries` metrics endpoint exposes E-Series metrics and requires the `target` parameter.

The `/metrics` endpoint exposes Go and process metrics for this exporter.

## Collectors

Collectors are enabled or disabled via a config file.

Name | Description | Default
-----|-------------|--------
drives | Collect status information about drives | Enabled
drive-statistics | Collect statistics on drives | Enabled
storage-systems | Collect status information about storage systems | Enabled
system-statistics | Collect storage system statistics | Enabled

## Configuration

The configuration defines targets that are to be queried. Example:

```yaml
modules:
  default:
    user: monitor
    password: secret
    proxy_url: http://localhost:8080
    timeout: 10
  status-only:
    user: monitor
    password: secret
    proxy_url: http://localhost:8080
    timeout: 10
    collectors:
      - drives
      - storage-systems
```

This exporter could then be queried via one of these two commands below.  The `eseries2` target will only run the `drives` and `storage-systems` collectors.

```
curl http://localhost:9310/eseries?target=eseries1
curl http://localhost:9310/eseries?target=eseries2&module=status-only
```

If no `timeout` is defined the default is `10`. 

## Dependencies

This exporter expects to communicate with SANtricity Web Services Proxy API and that your storage controllers are already setup to be accessed through that API.

## Install

Download the [latest release](https://github.com/treydock/eseries_exporter/releases)

Add the user that will run `eseries_exporter`

```
groupadd -r eseries_exporter
useradd -r -d /var/lib/eseries_exporter -s /sbin/nologin -M -g eseries_exporter -M eseries_exporter
```

Install compiled binaries after extracting tar.gz from release page.

```
cp /tmp/eseries_exporter /usr/local/bin/eseries_exporter
```

Install the necessary dependencies, see [dependencies section](#dependencies)

Add the necessary config, see [configuration section](#configuration)

Add systemd unit file and start service. Modify the `ExecStart` with desired flags.

```
cp systemd/eseries_exporter.service /etc/systemd/system/eseries_exporter.service
systemctl daemon-reload
systemctl start eseries_exporter
```

## Build from source

To produce the `eseries_exporter` binary:

```
make build
```

Or

```
go get github.com/treydock/eseries_exporter
```

## Prometheus configs

The following example assumes this exporter is running on the Prometheus server and communicating to the remote E-Series API.

```yaml
- job_name: eseries
  metrics_path: /eseries
  static_configs:
  - targets:
    - eseries1
    - eseries2
  relabel_configs:
  - source_labels: [__address__]
    target_label: __param_target
  - source_labels: [__param_target]
    target_label: instance
  - target_label: __address__
    replacement: 127.0.0.1:9313
- job_name: eseries-metrics
  metrics_path: /metrics
  static_configs:
  - targets:
    - localhost:9313
```
