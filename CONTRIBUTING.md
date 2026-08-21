# Contributing

Thanks for helping improve Codex Responses Lite Rules for CPA.

## Scope

Keep this plugin narrow. It should select CPA's native Responses Lite behavior by rule; it should not grow its own provider executor, credential store, HTTP client, request translator, retry system, or stream parser.

## Development

1. Use Go 1.26 and a working C toolchain.
2. Add or update tests for every behavior change.
3. Run `make check`.
4. Run `make build` to confirm the `c-shared` plugin still builds.
5. Update `CHANGELOG.md` for user-visible changes.

Pull requests should explain the use case, the smallest proposed behavior change, and how it was verified against CPA.

## Releases

Maintainers create semantic-version tags such as `v0.1.0`. The release workflow builds the Linux amd64 library, packages it with the filename required by the CPA Plugin Store, and publishes `checksums.txt`.
