#!/usr/bin/env python3
"""Validate the CPA Plugin Store manifest (registry.json) before it is served.

The store rejects a malformed manifest with `plugin_manifest_invalid`, which is
much harder to debug from the CPA side than from here. Run this after every
`scripts/update-registry.sh` and in CI.

Usage: scripts/check-registry.py [registry.json]
"""

from __future__ import annotations

import json
import re
import sys

PLUGIN_ID = "aq-codex-responses-lite"
SUPPORTED_ARCHES = ("amd64", "arm64")
VERSION_RE = re.compile(r"^\d+\.\d+\.\d+$")
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")


def fail(message: str) -> None:
    print(f"registry check failed: {message}", file=sys.stderr)
    sys.exit(1)


def check_artifact(plugin: dict, artifact: object, seen_arches: set[str]) -> None:
    plugin_id = plugin.get("id")
    version = plugin.get("version")
    if not isinstance(artifact, dict):
        fail(f"{plugin_id}: artifact entries must be objects")

    arch = artifact.get("goarch")
    if artifact.get("goos") != "linux":
        fail(f"{plugin_id}: only linux artifacts are published, got {artifact.get('goos')!r}")
    if arch not in SUPPORTED_ARCHES:
        fail(f"{plugin_id}: unsupported goarch {arch!r}, expected one of {', '.join(SUPPORTED_ARCHES)}")
    if arch in seen_arches:
        fail(f"{plugin_id}: duplicate linux/{arch} artifact")
    seen_arches.add(arch)

    expected_name = f"{plugin_id}_{version}_linux_{arch}.zip"
    expected_url = (
        f"https://github.com/{plugin.get('repository', '').removeprefix('https://github.com/')}"
        f"/releases/download/v{version}/{expected_name}"
    )
    url = artifact.get("url")
    if not isinstance(url, str) or not url.startswith("https://"):
        fail(f"{plugin_id}: {arch} url must be an https URL")
    if not url.endswith("/" + expected_name) or f"/download/v{version}/" not in url:
        fail(f"{plugin_id}: {arch} url must point at {expected_name} of v{version}, got {url}")
    if url != expected_url:
        fail(f"{plugin_id}: {arch} url must be {expected_url}, got {url}")

    sha256 = artifact.get("sha256")
    if not isinstance(sha256, str) or not SHA256_RE.match(sha256):
        fail(f"{plugin_id}: {arch} sha256 must be 64 lowercase hex characters")


def main() -> None:
    path = sys.argv[1] if len(sys.argv) > 1 else "registry.json"
    try:
        with open(path, encoding="utf-8") as handle:
            document = json.load(handle)
    except FileNotFoundError:
        fail(f"{path} does not exist")
    except json.JSONDecodeError as error:
        fail(f"{path} is not valid JSON: {error}")

    if document.get("schema_version") != 2:
        fail(f"schema_version must be 2, got {document.get('schema_version')!r}")

    plugins = document.get("plugins")
    if not isinstance(plugins, list) or not plugins:
        fail("plugins must be a non-empty list")

    ids: set[str] = set()
    for plugin in plugins:
        if not isinstance(plugin, dict):
            fail("each plugin entry must be an object")
        plugin_id = plugin.get("id")
        if not isinstance(plugin_id, str) or not plugin_id:
            fail("every plugin entry needs a non-empty id")
        if plugin_id in ids:
            fail(f"duplicate plugin id {plugin_id}")
        ids.add(plugin_id)
        if plugin_id != PLUGIN_ID:
            fail(f"unexpected plugin id {plugin_id!r}, expected {PLUGIN_ID!r}")

        for field in ("name", "description", "author", "repository", "license"):
            if not isinstance(plugin.get(field), str) or not plugin[field].strip():
                fail(f"{plugin_id}: {field} must be a non-empty string")

        version = plugin.get("version")
        if not isinstance(version, str) or not VERSION_RE.match(version):
            fail(f"{plugin_id}: version must look like x.y.z, got {version!r}")

        install = plugin.get("install")
        if not isinstance(install, dict) or install.get("type") != "direct":
            fail(f"{plugin_id}: install.type must be 'direct'")
        artifacts = install.get("artifacts")
        if not isinstance(artifacts, list) or not artifacts:
            fail(f"{plugin_id}: install.artifacts must be a non-empty list")

        seen_arches: set[str] = set()
        for artifact in artifacts:
            check_artifact(plugin, artifact, seen_arches)

    print(f"{path}: ok ({len(plugins)} plugin, schema v2, direct artifacts)")


if __name__ == "__main__":
    main()
