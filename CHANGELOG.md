# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-04-27

### Added
- **Native Environment Variable Support**: Placeholders like `${VAR}` or `$VAR` in `config.yaml` and `data.yaml` are now automatically expanded using system environment variables.
- **Resource ID Prefixing**: Added the `id_prefix` configuration key to prepend a namespace (e.g., `user_service.`) to all synchronized resource IDs.
- **Scoped Cleanup**: When `id_prefix` is set, `reset_on_start` will only delete resources matching that prefix, enabling multi-tenant APISIX usage.

### Changed
- Improved `reset_on_start` logic to be much safer and targeted when using prefixes.
- Updated example configuration and data files to demonstrate new features.
- Added comprehensive unit tests for configuration loading and ID qualification.
