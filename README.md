# Codex Responses Lite Rules for CPA

[![CI](https://github.com/AstroQore/cpa-plugin-codex-responses-lite/actions/workflows/ci.yml/badge.svg)](https://github.com/AstroQore/cpa-plugin-codex-responses-lite/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/AstroQore/cpa-plugin-codex-responses-lite)](https://github.com/AstroQore/cpa-plugin-codex-responses-lite/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A small [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) request-interceptor plugin that enables CPA's native **Codex Responses Lite** path for selected provider-prefix/model pairs.

[简体中文](README.zh-CN.md)

## Why this plugin exists

CPA's Codex executor can automatically add the hosted `image_generation` tool. Some OpenAI-compatible upstream models support the Responses API but do not support that hosted tool. CPA already has a Responses Lite path that skips the automatic injection; this plugin selects that existing path with a narrowly scoped rule.

```mermaid
flowchart LR
    A[Client request] --> B{Exact provider/model match?}
    B -- No --> C[Leave request unchanged]
    B -- Yes --> D[Add Responses Lite header]
    D --> E[CPA native Codex executor]
    C --> E
    E --> F[Existing provider, auth, proxy, and stream handling]
```

The plugin adds one internal request header when a rule matches:

```http
X-OpenAI-Internal-Codex-Responses-Lite: true
```

Everything after that remains native CPA behavior.

## Features

- Exact, case-sensitive matching on `provider_prefix/model`.
- Multiple provider prefixes and model allowlists.
- Applies both before and after credential selection.
- No network client, credentials, upstream URL, proxy, retry loop, stream parser, or usage accounting.
- No request-body rewriting.
- Fails configuration early for empty, duplicate, or ambiguous rules.

## Requirements

- CLIProxyAPI with dynamic plugin support. The initial release is tested with CPA `v7.2.127`.
- A target model already configured on CPA's native Codex/Responses-capable path, normally under `codex-api-key`.
- Provider prefixes enabled in CPA so clients can request `prefix/model`.
- Prebuilt Linux amd64 and Linux arm64 artifacts. Other platforms can build from source.

This plugin does **not** register models or move models between provider sections. Configure models and credentials in CPA first, then use this plugin only to select Responses Lite for the desired prefixed model names.

## Installation

1. Download the release asset for your platform and verify it against `checksums.txt`.
2. Put the dynamic library in CPA's plugin directory. The filename must remain `aq-codex-responses-lite.so` on Linux, `.dylib` on macOS, or `.dll` on Windows.
3. Add the plugin configuration to CPA's `config.yaml`.
4. Restart or reload CPA according to your deployment process, then verify registration in the CPA logs or plugin management API.

Example for Linux amd64:

```bash
unzip aq-codex-responses-lite_0.1.0_linux_amd64.zip
install -m 0755 aq-codex-responses-lite.so /path/to/cpa/plugins/
```

Example for Linux arm64 (for example Raspberry Pi 64-bit, ARM servers, Apple Silicon VMs):

```bash
unzip aq-codex-responses-lite_0.1.0_linux_arm64.zip
install -m 0755 aq-codex-responses-lite.so /path/to/cpa/plugins/
```

## Configuration

```yaml
plugins:
  enabled: true
  dir: /path/to/cpa/plugins
  configs:
    aq-codex-responses-lite:
      enabled: true
      priority: 200
      rules:
        - provider_prefix: opencode
          models:
            - grok-4.5
            - muse-spark-1.2-contributor
        - provider_prefix: another-provider
          models:
            - another-model
```

See [examples/config.yaml](examples/config.yaml) for a copy-ready example.

### Matching behavior

| Requested model | Example rule | Result |
| --- | --- | --- |
| `opencode/grok-4.5` | prefix `opencode`, model `grok-4.5` | Responses Lite enabled |
| `grok-4.5` | same rule | Unchanged; bare names never match |
| `opencode/other-model` | same rule | Unchanged |
| `OpenCode/grok-4.5` | same rule | Unchanged; matching is case-sensitive |
| `other/grok-4.5` | same rule | Unchanged |

## Important behavior and limits

- Responses Lite also makes CPA set `parallel_tool_calls=false`. Test function calling as well as plain text for every newly enabled model.
- The plugin prevents CPA's **automatic** hosted `image_generation` injection for matching requests. It does not remove tools explicitly supplied by a client.
- It does not guarantee that an upstream model supports every other Responses API feature.
- It does not alter Chat Completions traffic or non-matching model requests.
- A configuration reload with invalid plugin rules is rejected instead of silently broadening the match.

## Build and test

Go `1.26` and a C toolchain are required because CPA plugins use Go's `c-shared` build mode.

```bash
make check
make build
```

The library is written to `dist/`. On a Linux amd64 host, maintainers can create the store-compatible amd64 release files with:

```bash
make package VERSION=0.1.0
```

To package for Linux arm64, install the cross toolchain (`gcc-aarch64-linux-gnu` on Debian/Ubuntu) and set the target architecture:

```bash
make package VERSION=0.1.0 TARGET_ARCH=arm64
```

The release zip contains exactly one root-level dynamic library, as required by the CPA Plugin Store.

## Troubleshooting

**The plugin loads but does not match**

Use the exact client-visible prefixed model name. Bare model names are intentionally ignored. Prefix and model matching are case-sensitive.

**The request still contains `image_generation`**

Check whether the tool was supplied by the client. This plugin only disables CPA's automatic hosted-tool injection. Also confirm that CPA registered the plugin as a request interceptor and loaded the expected rule.

**Tool calls behave differently**

Responses Lite disables parallel tool calls in CPA. Prefer `tool_choice=auto` unless the upstream explicitly supports stricter modes, and validate the upstream model directly.

## Security

Plugin configuration contains only provider prefixes and model identifiers. Do not put API keys or upstream credentials in it. See [SECURITY.md](SECURITY.md) for vulnerability reporting.

## Contributing

Issues and focused pull requests are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting changes.

## License

[MIT](LICENSE) © AstroQore.
