# Changelog

All notable changes to this project are documented in this file.

## [0.1.1] - 2026-09-22

### Added

- Linux arm64 release packaging. Release builds now publish both `linux_amd64` and `linux_arm64` zips, and `make package` supports `TARGET_ARCH=arm64` cross builds.

## [0.1.0] - 2026-08-21

### Added

- Exact provider-prefix/model rules for selecting CPA's native Codex Responses Lite path.
- Request interception before and after credential selection.
- Configuration validation and matching tests.
- Linux amd64 release packaging for the CPA Plugin Store.

[0.1.0]: https://github.com/AstroQore/cpa-plugin-codex-responses-lite/releases/tag/v0.1.0

[0.1.1]: https://github.com/lzy-xkwh/cpa-plugin-codex-responses-lite/releases/tag/v0.1.1
