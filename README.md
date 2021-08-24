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
drive-statistics | Collect statistics on drives | Disabled
controller-statistics | Collect controller statistics | Enabled
storage-systems | Collect status information about storage systems | Enabled
system-statistics | Collect storage system statistics | Enabled
hardware-inventory | Collect hardware inventory statuses | Enabled

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
  ssl:
    user: monitor
    password: secret
    proxy_url: https://proxy.example.com
    root_ca: /etc/pki/tls/root.pem
    insecure_ssl: false
    timeout: 10
```

This exporter could then be queried via one of these two commands below.  The `eseries2` target will only run the `drives` and `storage-systems` collectors.

```
curl http://localhost:9310/eseries?target=eseries1
curl http://localhost:9310/eseries?target=eseries2&module=status-only
```

If no `timeout` is defined the default is `10`.

If the HTTP schema used for `proxy_url` is `https` then the exporter will attempt to use the system CA truststore as well as any root CA specified with `root_ca` option.  By default certificate verification is enabled, set `insecure_ssl` to disable SSL verification.

## Dependencies

This exporter expects to communicate with SANtricity Web Services Proxy API and that your storage controllers are already setup to be accessed through that API.

This repo provides a Docker based approach to running the Web Services Proxy:

```
cd webservices_proxy
docker build -t webservices_proxy .
```

The above Docker container will have `admin:admin` as the credentials and can be run using a command like the following:

```
docker run -d --rm -it --name webservices_proxy --network host -e ACCEPT_EULA=true \
-v /var/lib/eseries_webservices_proxy/working:/opt/netapp/webservices_proxy/working \
webservices_proxy
```

**NOTE**: During testing it seemed in order for a Docker based proxy to communicate with E-Series controllers the container had to use the host's network.

Example of settig up the Web Services Proxy with an E-Series system.  Replace `PASSWORD` with the password for your E-Series system.  Replace `ID` with the name of your system.  With `IP1` and `IP2` with IP addresses of your controllers for the system.

```
curl -X POST -u admin:admin "http://localhost:8080/devmgr/v2/storage-systems" \
-H  "accept: application/json" -H  "Content-Type: application/json" \
-d '{
  "id": "ID", "controllerAddresses": [ "IP1" , "IP2" ],
  "acceptCertificate": true, "validate": false, "password": "PASSWORD"
}'
```

If your storage systems have UUID style IDs this is a way to query the names for each ID:

```
$ curl -u admin:admin http://localhost:8080/devmgr/v2/storage-systems 2>/dev/null | \
  jq -r '.[] | "ID: \(.id)\tname: \(.name)"'
ID: f0d2fadc-3e16-46c5-b62e-c9ab6d430b50    name: eseries1
ID: 25db8d36-6732-495d-b693-8add202750d6    name: eseries2
```

## Docker

Example of running the Docker container

```
docker run -d -p 9313:9313 -v "eseries_exporter.yaml:/eseries_exporter.yaml:ro" treydock/eseries_exporter
```

This repo also provides a Docker Compose file that can be used to run both the Web Services Proxy and this exporter.

```
docker-compose up -d
```

See [dependencies section](#dependencies) for steps necessary to bootstrap the Web Services Proxy.

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

The following is an example if your E-Series web proxy is using UUIDs or other cryptic IDs:

```yaml
- job_name: eseries
  metrics_path: /eseries
  static_configs:
  - targets:
    - f0d2fadc-3e16-46c5-b62e-c9ab6d430b50
    labels:
      name: eseries1
  - targets:
    - 25db8d36-6732-495d-b693-8add202750d6
    labels:
      name: eseries2
  relabel_configs:
  - source_labels: [__address__]
    target_label: __param_target
  - target_label: __address__
    replacement: 127.0.0.1:9313
  - source_labels: [name]
    target_label: instance
  - regex: '^(name)$'
    action: labeldrop
```
