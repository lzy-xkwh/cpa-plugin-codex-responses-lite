# Contributing

Thanks for helping improve Codex Responses Lite Rules for CPA.

## Scope

Keep this plugin narrow. It should select CPA's native Responses Lite behavior by rule; it should not grow its own provider executor, credential store, HTTP client, request translator, retry system, or stream parser.

## Development

1. Use Go 1.26 and a working C toolchain.
2. Add or update tests for every behavior change.
3. Run `make check`.
4. Run `make build` to confirm the `c-shared` plugin still builds.
5. Run `make check-registry` if you touched `registry.json`.
6. Update `CHANGELOG.md` for user-visible changes.

Pull requests should explain the use case, the smallest proposed behavior change, and how it was verified against CPA.

## Releases

Maintainers create semantic-version tags such as `v0.3.0`; bump `pluginVersion` in
`main.go` to the tag first. The release workflow then:

1. builds the Linux amd64 and arm64 libraries, packaging each as
   `aq-codex-responses-lite_<version>_linux_<arch>.zip` — the filename layout the
   CPA Plugin Store expects;
2. publishes both zips plus `checksums.txt`;
3. runs `scripts/update-registry.sh <version> --artifacts-dir dist`, validates the
   manifest with `scripts/check-registry.py`, and pushes the refreshed
   `registry.json` to `main`, so the store serves the new version.

If step 3 fails (for example `main` moved while the release was building), finish it
by hand:

```bash
make registry VERSION=0.3.0
make check-registry
git commit registry.json -m "chore: point plugin store registry at v0.3.0"
git push origin main
```

`registry.json` on `main` always points at the newest published release. A tag
therefore carries the pointer of the version released before it, which is expected:
the store reads the manifest from `main`, never from a tag.
