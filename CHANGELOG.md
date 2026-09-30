# Changelog

All notable changes to this project are documented in this file.

## [0.3.0] - 2026-09-30

### Added

- CPA Plugin Store manifest (`registry.json`, schema v2 with `direct` linux amd64/arm64 artifacts) plus README instructions for installing and updating the plugin from the store.
- `scripts/update-registry.sh` and `scripts/check-registry.py`, exposed as `make registry VERSION=x.y.z` and `make check-registry`.
- The release workflow refreshes `registry.json` for every published tag and pushes it to `main`, so store installs follow new releases without a manual copy.
- CI validates `registry.json`, so a malformed manifest cannot reach the store.

### Changed

- `make package` now requires an explicit `VERSION=` instead of defaulting to `0.1.0`.

## [0.2.0] - 2026-09-24

### Added

- Management-UI config wizard. The plugin registers a `Responses Lite` resource page in the CPA admin panel (`/v0/resource/plugins/aq-codex-responses-lite/config-wizard`) for editing `rules` visually; saves go through the CPA management API (`PUT /v0/management/plugins/aq-codex-responses-lite/config`) with the admin's own key.

## [0.1.2] - 2026-09-24

### Changed

- Missing or empty `rules` no longer fails plugin reconfiguration. The plugin loads inactive (matches nothing) so store installs work before any rule is configured; non-empty rules still fail fast for missing, duplicate, or ambiguous entries.

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

[0.2.0]: https://github.com/lzy-xkwh/cpa-plugin-codex-responses-lite/releases/tag/v0.2.0

[0.3.0]: https://github.com/lzy-xkwh/cpa-plugin-codex-responses-lite/releases/tag/v0.3.0
