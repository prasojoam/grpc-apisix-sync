# grpc-apisix-sync

A tool to synchronize gRPC services with Apache APISIX.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Introduction

I've worked on a project that uses gRPC for inter-service communication. The project uses APISIX as an API Gateway. It was a pain to manually create routes and services in APISIX for each gRPC service and methods. Let alone maintaining production and development environments. For that reason, I decided to create this tool to automate the process.

## ✨ Features

- **Service Synchronization**: Synchronize gRPC services with APISIX.
- **Route Synchronization**: Synchronize gRPC routes with APISIX.

## 🛠 Installation

### From Source
```bash
go install github.com/prasojoam/grpc-apisix-sync@latest
```

## 🚀 Usage

### Basic Sync
Synchronize a proto file with APISIX:
```bash
grpc-apisix-sync --config ./config.yaml --data ./data.yaml
```

## ⚙️ Configuration

You can use a configuration file (`config.yaml`) to manage settings:

```yaml
apisix:
  url: "http://localhost:9180"
  key: "your-apisix-key "
proto:
  path: "./proto"
  includes:
    - "/usr/local/include"
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
