# Quota Management Service

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![test](https://github.com/Blinkuu/qms/actions/workflows/test.yml/badge.svg)](https://github.com/Blinkuu/qms/actions/workflows/test.yml)
[![lint](https://github.com/Blinkuu/qms/actions/workflows/lint.yml/badge.svg)](https://github.com/Blinkuu/qms/actions/workflows/lint.yml)
[![go report](https://goreportcard.com/badge/github.com/Blinkuu/qms)](https://goreportcard.com/report/github.com/Blinkuu/qms)
[![codecov](https://codecov.io/gh/Blinkuu/qms/branch/main/graph/badge.svg?token=73nMinG7d9)](https://codecov.io/gh/Blinkuu/qms)

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Deployment](#deployment)
    - [Monolithic mode](#monolithic-mode)
    - [Microservices mode](#microservices-mode)
- [API](#api)
    - [Ping](#ping)
    - [Ready](#ready)
    - [Health](#health)
    - [Metrics](#metrics)
    - [Memberlist](#memberlist)
    - [Allow](#allow)
    - [View](#view)
    - [alloc](#alloc)
    - [free](#free)

## Overview

Quota Management Service (QMS) is a generic, horizontally-scalable, highly-available, and fault-tolerant system for
managing quotas. QMS is different from other quota systems in that it natively supports both allocation and rate quotas.

Rate quotas are typically used to limit the number of requests to a shared resource, such as an API or service. Their
key characteristic is that they reset after a specified time interval. QMS provides extensible support for this type of
workload and can be configured to use various rate-limiting algorithms. Currently, the rate component
supports only the `memory` storage backend.

Allocation quotas are commonly used to restrict the use of resources that do not have a usage rate. Common examples
include limiting the amount of used cloud storage or instances deployed. An essential property of allocation quotas is
that they do not reset over time and must be explicitly released when they are no longer needed.

> Warning: QMS is not production ready. This project was developed as part of my Master's Thesis and requires further
> polishing to consider it stable.

## Quick Start

This repository hosts all information needed to build and run QMS from source. There are two options to build and run
QMS locally:

#### You have a working Go environment

```bash
# Clone QMS repository from GitHub
git clone https://github.com/Blinkuu/qms

# Change directory
cd qms

# Build QMS
make build

# Run QMS locally in a monolithic mode
./bin/qms
```

#### You have a working Tilt environment

```bash
# Clone QMS repository from GitHub
git clone https://github.com/Blinkuu/qms

# Change directory
cd qms

# Run QMS locally in microservices mode
tilt up microservices
```

## Configuration

Quota Management Service leverages the YAML file format for configuration. Below is an example of the most basic
configuration file that can be used to run QMS in monolithic mode on a local machine. For more advanced configuration
examples, please refer to the `
configs` directory.

```yaml
target: all
otel_collector_target: agent:4317

server:
  http_port: 6789

memberlist:
  join_addresses:
    - 127.0.0.1:7946

proxy:
  alloc_addresses:
    - 127.0.0.1:6789
  rate_addresses:
    - 127.0.0.1:6789

rate:
  storage:
    backend: memory
  quotas:
    - namespace: namespace1
      resource: resource1
      strategy:
        algorithm: token-bucket
        unit: minute
        requests_per_unit: 120

alloc:
  storage:
    backend: memory
  quotas:
    - namespace: namespace1
      resource: resource1
      strategy:
        capacity: 10
```

## Deployment

QMS has a microservices-based architecture and is designed to run as a horizontally scalable distributed system. There
are three primary components: QMS Proxy, QMS Rate, and QMS Alloc. All those components can run
separately and in parallel. Because of QMS’s architecture and implementation, all of the
code compiles into a single binary. The `target` option controls the behavior of this single
binary at runtime and determines which components will be run.

### Monolithic mode

The monolithic mode runs all required components in a single process. This is the default mode and can be defined by
specifying `-target=all` command-line flag or configuring the target parameter in the YAML config file. Monolithic mode
is the simplest mode of operation, and it is useful for getting started quickly to experiment with QMS or setting up the
system in the development environment.

### Microservices mode

In microservices deployment mode, components are deployed as separate processes. Similarly to the monolithic mode,
the `target` option specifies which component is run. There are three possible choices: `proxy`, `rate`, and `alloc`.
Microservices mode is the recommended method for production deployment as it offers the greatest level of flexibility
and control over failure domains.

In microservices mode, scaling is done on a per-component basis. As a result, each component can be scaled independently
according to the needs. This flexibility comes at a cost. Microservices mode is a bit more complex to set up, deploy and
operate than the monolithic mode. The recommended way to run QMS in this mode is to use Kubernetes.

## API

QMS by default exposes a JSON over HTTP API.

### Ping

Pings the instance. Can be used to check basic availability. For a more advanced liveness check, please refer to
the [health](#health) endpoint.

```
GET /ping
```

**Example response**

```json
{
  "status": 1001,
  "msg": "ok",
  "result": {
    "msg": "pong"
  }
}
```

### Ready

Check whether an instance is ready to accept traffic. This endpoint is designed to work with the Kubernetes readiness
probe.

```
GET /ready
```

**Example response**

```json
{
  "status": 1001,
  "msg": "ok"
}
```

### Health

Check whether an instance is in a healthy state. This endpoint is designed to work with the Kubernetes liveness
probe.

```
GET /healthz
```

**Example response**

```json
{
  "status": 1001,
  "msg": "ok"
}
```

### Metrics

Gets instance metrics in a Prometheus-compatible format.

```
GET /metrics
```

For response format, please refer to the Prometheus [documentation](https://prometheus.io/docs/introduction/overview/).

### Memberlist

Returns the current view of the cluster as seen by an instance.

```
GET /memberlist
```

**Example response**

```json
{
  "status": 1001,
  "msg": "ok",
  "result": {
    "members": [
      {
        "service": "proxy",
        "hostname": "qms-proxy-5bfc6ccf44-tmwd9",
        "host": "10.1.54.225",
        "http_port": 6789,
        "gossip_port": 7946
      },
      {
        "service": "proxy",
        "hostname": "qms-proxy-5bfc6ccf44-bpjs2",
        "host": "10.1.54.224",
        "http_port": 6789,
        "gossip_port": 7946
      },
      {
        "service": "proxy",
        "hostname": "qms-proxy-5bfc6ccf44-7jqml",
        "host": "10.1.54.222",
        "http_port": 6789,
        "gossip_port": 7946
      }
    ]
  }
}
```

### Allow

Checks whether a request to a particular resource can be allowed based on the definition of a concrete rate quota. This
is the endpoint that can be used to implement rate limiting with QMS.

```
POST /api/v1/allow
```

**Parameters**

|   Name    |  Type  |  In  |              Description              |
|:---------:|:------:|:----:|:-------------------------------------:|
| namespace | string | body | Namespace where the resource resides. |
| resource  | string | body |         Name of the resource.         |
|  tokens   |  int   | body |     Amount of tokens to request.      |

**Example response**

```json
{
  "status": 1001,
  "msg": "ok",
  "result": {
    "wait_time": 0,
    "ok": true
  }
}
```

### View

Returns the current status of a particular allocation quota.

```
POST /api/v1/view
```

**Parameters**

|   Name    |  Type  |  In  |              Description              |
|:---------:|:------:|:----:|:-------------------------------------:|
| namespace | string | body | Namespace where the resource resides. |
| resource  | string | body |         Name of the resource.         |

**Example response**

```json
{
  "status": 1001,
  "msg": "ok",
  "result": {
    "allocated": 14,
    "capacity": 100,
    "version": 3
  }
}
```

### Alloc

Acquires a certain amount of tokens from a particular allocation quota.

```
POST /api/v1/alloc
```

**Parameters**

|   Name    |  Type  |  In  |                                             Description                                             |
|:---------:|:------:|:----:|:---------------------------------------------------------------------------------------------------:|
| namespace | string | body |                                Namespace where the resource resides.                                |
| resource  | string | body |                                        Name of the resource.                                        |
|  tokens   |  int   | body |                                    Amount of tokens to request.                                     |
|  version  |  int   | body | Current version of the resource. If set to 0, no optimistic concurrency control check is performed. |

```json
{
  "msg": "ok",
  "result": {
    "remaining_tokens": 86,
    "current_version": 3,
    "ok": true
  }
}
```

### Free

Releases a certain amount of tokens from a particular allocation quota.

```
POST /api/v1/free
```

**Parameters**

|   Name    |  Type  |  In  |                                             Description                                             |
|:---------:|:------:|:----:|:---------------------------------------------------------------------------------------------------:|
| namespace | string | body |                                Namespace where the resource resides.                                |
| resource  | string | body |                                        Name of the resource.                                        |
|  tokens   |  int   | body |                                    Amount of tokens to request.                                     |
|  version  |  int   | body | Current version of the resource. If set to 0, no optimistic concurrency control check is performed. |

**Example response**

```json
{
  "status": 1001,
  "msg": "ok",
  "result": {
    "remaining_tokens": 87,
    "current_version": 4,
    "ok": true
  }
}
```

## Contributing

Contributions are very welcome! Either by reporting issues or submitting pull requests.

## License

QMS is distributed under the [AGPL-3.0 license](https://github.com/Blinkuu/qms/blob/main/LICENSE.md).
