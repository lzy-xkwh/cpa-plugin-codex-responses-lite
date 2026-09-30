#!/usr/bin/env bash
# Point registry.json (the CPA Plugin Store manifest) at a published release:
# plugin version, artifact URLs and SHA-256 digests.
#
# Usage:
#   scripts/update-registry.sh <version> [--artifacts-dir DIR] [--registry FILE] [--repo OWNER/NAME]
#
#   <version>          "0.3.0" or "v0.3.0"
#   --artifacts-dir    directory holding the release zips; their SHA-256 is
#                      computed locally instead of downloaded (this is how the
#                      release workflow runs it)
#   --registry         manifest to update (default: registry.json)
#   --repo             GitHub repository holding the release
#                      (default: lzy-xkwh/cpa-plugin-codex-responses-lite)
#
# Without --artifacts-dir the digests are read from the release's checksums.txt,
# so the manifest can also be refreshed from a maintainer machine after a
# release has been published.
set -euo pipefail

plugin_id="aq-codex-responses-lite"
arches="amd64 arm64"
repo="lzy-xkwh/cpa-plugin-codex-responses-lite"
registry_file="registry.json"
artifacts_dir=""

usage() {
	cat <<'USAGE'
usage: scripts/update-registry.sh <version> [--artifacts-dir DIR] [--registry FILE] [--repo OWNER/NAME]
USAGE
}

version=""
if [ $# -gt 0 ]; then
	case "$1" in
		-*) ;;
		*) version="$1"; shift ;;
	esac
fi

while [ $# -gt 0 ]; do
	case "$1" in
		--artifacts-dir) artifacts_dir="${2:?--artifacts-dir needs a value}"; shift 2 ;;
		--registry) registry_file="${2:?--registry needs a value}"; shift 2 ;;
		--repo) repo="${2:?--repo needs a value}"; shift 2 ;;
		-h|--help) usage; exit 0 ;;
		*) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
	esac
done

version="${version#v}"
if [ -z "$version" ]; then
	echo "error: a release version is required" >&2
	usage >&2
	exit 2
fi
case "$version" in
	[0-9]*.[0-9]*.[0-9]*) ;;
	*) echo "error: version must look like x.y.z, got '$version'" >&2; exit 2 ;;
esac
if [ ! -f "$registry_file" ]; then
	echo "error: $registry_file not found (run from the repository root)" >&2
	exit 1
fi

checksums="$(mktemp)"
trap 'rm -f "$checksums"' EXIT

if [ -n "$artifacts_dir" ]; then
	if [ ! -d "$artifacts_dir" ]; then
		echo "error: --artifacts-dir '$artifacts_dir' is not a directory" >&2
		exit 1
	fi
	for arch in $arches; do
		zip_name="${plugin_id}_${version}_linux_${arch}.zip"
		if [ ! -f "$artifacts_dir/$zip_name" ]; then
			echo "error: expected release asset $artifacts_dir/$zip_name" >&2
			exit 1
		fi
		(cd "$artifacts_dir" && sha256sum "$zip_name") >> "$checksums"
	done
	echo "checksums: computed from $artifacts_dir"
else
	url="https://github.com/${repo}/releases/download/v${version}/checksums.txt"
	echo "checksums: downloading $url"
	curl -fsSL "$url" -o "$checksums" || {
		echo "error: could not download $url" >&2
		exit 1
	}
fi

python3 - "$registry_file" "$version" "$repo" "$plugin_id" "$arches" "$checksums" <<'PY'
import json
import sys

registry_path, version, repo, plugin_id, arches, checksums_path = sys.argv[1:7]

digests = {}
with open(checksums_path, encoding="utf-8") as handle:
    for line in handle:
        parts = line.split()
        if len(parts) < 2:
            continue
        digests[parts[-1].lstrip("*")] = parts[0]

with open(registry_path, encoding="utf-8") as handle:
    document = json.load(handle)

if document.get("schema_version") != 2:
    sys.exit(f"error: {registry_path} must use schema_version 2")

matches = [plugin for plugin in document.get("plugins", []) if plugin.get("id") == plugin_id]
if len(matches) != 1:
    sys.exit(f"error: expected exactly one plugin entry with id {plugin_id}, found {len(matches)}")
plugin = matches[0]
plugin["version"] = version

install = plugin.setdefault("install", {})
if install.get("type") != "direct":
    sys.exit("error: install.type must be 'direct'")
artifacts = install.setdefault("artifacts", [])

for arch in arches.split():
    zip_name = f"{plugin_id}_{version}_linux_{arch}.zip"
    digest = digests.get(zip_name)
    if not digest or len(digest) != 64:
        sys.exit(f"error: no SHA-256 for {zip_name} in the checksums source")
    url = f"https://github.com/{repo}/releases/download/v{version}/{zip_name}"

    artifact = next(
        (item for item in artifacts if item.get("goos") == "linux" and item.get("goarch") == arch),
        None,
    )
    if artifact is None:
        artifact = {"goos": "linux", "goarch": arch}
        artifacts.append(artifact)
    artifact["url"] = url
    artifact["sha256"] = digest

with open(registry_path, "w", encoding="utf-8") as handle:
    json.dump(document, handle, indent=2, ensure_ascii=False)
    handle.write("\n")

print(f"{registry_path}: {plugin_id} v{version} ({', '.join(arches.split())})")
PY

python3 "$(dirname "$0")/check-registry.py" "$registry_file"
