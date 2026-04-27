# grpc-apisix-sync

A tool to synchronize gRPC services with Apache APISIX.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Introduction

I've worked on a project that uses gRPC for inter-service communication. The project uses APISIX as an API Gateway. It was a pain to manually create routes and services in APISIX for each gRPC service and methods. Let alone maintaining production and development environments. For that reason, I decided to create this tool to automate the process.

## ✨ Features

- **Service Synchronization**: Synchronize gRPC services with APISIX.
- **Route Synchronization**: Synchronize gRPC routes with APISIX.
- **Zero-Dependency Proto Compilation**: Pure-Go proto compiler—no `protoc` binary required!
- **Parallelized Cleanup**: Optional `reset_on_start` to wipe the gateway state before syncing.
- **Native Environment Variable Support**: Use `${VAR}` syntax in YAML files for dynamic configuration.

## 🛠 Installation

### From Source
```bash
go install github.com/prasojoam/grpc-apisix-sync@latest
```

## 🚀 Usage

### Command
```bash
grpc-apisix-sync --config ./config.yaml --data ./data.yaml
```

## 🌍 Environment Variables

The tool natively supports environment variable expansion in both `config.yaml` and `data.yaml`. This allows you to keep sensitive information out of your version control and easily switch between environments.

### Syntax
Use `${VARIABLE_NAME}` or `$VARIABLE_NAME` anywhere in your YAML files.

### Example
```yaml
# data.yaml
upstreams:
  - id: "auth-service"
    nodes:
      - host: "${AUTH_SERVICE_HOST}"
        port: ${AUTH_SERVICE_PORT}
```

```bash
export AUTH_SERVICE_HOST="auth.production.svc"
export AUTH_SERVICE_PORT=8080
grpc-apisix-sync --data data.yaml
```

## ⚙️ Configuration (`config.yaml`)

| Field | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `apisix.url` | string | **Yes** | - | The base URL of the APISIX Admin API. |
| `apisix.key` | string | **Yes** | - | The API key for the APISIX Admin API. |
| `proto.includes` | string[] | No | `[]` | Additional search paths for `.proto` imports. |
| `reset_on_start` | boolean | No | `false` | If true, wipes existing Routes, Services, Upstreams, and Protos before syncing. |

### Example `config.yaml`
```yaml
apisix:
  url: "http://localhost:9180"
  key: "your-apisix-key "
proto:
  includes:
    - "/usr/local/include"
reset_on_start: false
```

## 📊 Data Mapping (`data.yaml`)

The `data.yaml` defines the infrastructure you want to sync.

### Protos
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | string | **Yes** | The ID used to register the proto in APISIX. |
| `path` | string | **Yes** | The local filesystem path to the `.proto` file. |

### Upstreams
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | string | **Yes** | The ID used to register the upstream in APISIX. |
| `nodes` | object[] | **Yes** | List of nodes for the upstream. |
| `nodes[].host` | string | **Yes** | The host (IP or domain) of the gRPC server. |
| `nodes[].port` | int | **Yes** | The port of the gRPC server. |
| `nodes[].weight` | int | No (1) | Load balancing weight. |

### Services
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | string | **Yes** | The ID used to register the service in APISIX. |
| `upstream` | string | **Yes** | The ID of the upstream defined in the same file. |

### Route Defaults
Optional section to reduce repetition in the `routes` list.
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `service` | string | No | Fallback service ID for routes without one. |
| `proto` | string | No | Fallback proto ID for routes without one. |

### Routes
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | string | **Yes** | The ID used to register the route in APISIX. |
| `uri` | string | **Yes** | The public HTTP URI. |
| `service` | string | No | Overrides `route_defaults.service`. |
| `proto` | string | No | Overrides `route_defaults.proto`. |
| `methods` | string | No | Comma-separated HTTP methods (e.g. `POST,OPTIONS`). |
| `grpc` | string | **Yes** | The gRPC method in `package.Service/Method` format. |

### Example `data.yaml`
```yaml
protos:
  - id: "proto.user_service"
    path: "./proto/user_service.proto"

upstreams:
  - id: "upstream.user_service"
    nodes:
      - host: "localhost"
        port: 50051

services:
  - id: "service.user_service"
    upstream: "upstream.user_service"

route_defaults:
  service: "service.user_service"
  proto: "proto.user_service"

routes:
  - id: "route.user_service.login"
    uri: "/user/login"
    methods: "POST"
    grpc: "user.User/Login"
  - id: "route.user_service.get_profile"
    uri: "/user/profile"
    methods: "GET"
    grpc: "user.User/GetProfile"
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git checkout origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

Built with ❤️ by [prasojoam](https://github.com/prasojoam)
