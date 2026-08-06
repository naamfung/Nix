#!/usr/bin/env bash
# Verify the Windows portable release unit (versioned-v1) before packaging.
# This must run on the Windows staging directory: NTFS treats file names
# case-insensitively, so Inx.exe and inx.exe silently overwrite each
# other even though a source-level packaging test sees two different strings.
set -euo pipefail

staging="${1:?usage: verify-windows-portable.sh STAGING_DIR}"
[ -d "$staging" ] || { echo "Windows portable staging directory is missing: $staging" >&2; exit 1; }

# Root entry points for versioned-v1 (no Guard, no flat desktop/helper).
required_root=(
	"Inx.exe"
	"inx-cli.exe"
	"inx-launcher.exe"
	"current.json"
)

for expected in "${required_root[@]}"; do
	[ -e "$staging/$expected" ] || {
		echo "Windows portable root entry is missing: $expected" >&2
		exit 1
	}
done

# Exactly one active version directory under versions/v*.
version_dirs=()
if [ -d "$staging/versions" ]; then
	while IFS= read -r -d '' d; do
		version_dirs+=("$d")
	done < <(find "$staging/versions" -mindepth 1 -maxdepth 1 -type d -name 'v*' -print0 2>/dev/null || true)
fi
if [ "${#version_dirs[@]}" -lt 1 ]; then
	echo "Windows portable versions/ tree is missing an active version directory" >&2
	exit 1
fi

# current.json must point at a real version directory.
active_dir=$(python3 - <<'PY' "$staging/current.json" 2>/dev/null || true
import json,sys
p=sys.argv[1]
with open(p) as f:
    d=json.load(f)
print(d.get("activeDir",""))
PY
)
if [ -z "$active_dir" ]; then
	# Fallback without python: require versions/v* with required members
	active_dir=""
	for d in "${version_dirs[@]}"; do
		base=$(basename "$d")
		if [ -f "$d/inx-desktop.exe" ] && [ -f "$d/inx-cli.exe" ] && [ -f "$d/inx-update-helper.exe" ]; then
			active_dir="versions/$base"
			break
		fi
	done
fi
[ -n "$active_dir" ] || {
	echo "Windows portable current.json activeDir is empty or unreadable" >&2
	exit 1
}
version_path="$staging/$active_dir"
[ -d "$version_path" ] || {
	echo "Windows portable activeDir does not exist: $active_dir" >&2
	exit 1
}
for name in inx-desktop.exe inx-cli.exe inx-update-helper.exe; do
	[ -f "$version_path/$name" ] || {
		echo "Windows portable version member is missing: $active_dir/$name" >&2
		exit 1
	}
done

# Guard must not persist in a normal portable layout.
if [ -e "$staging/inx-guard.exe" ] || [ -e "$staging/inx-desktop.exe" ]; then
	echo "Windows portable must not ship flat inx-guard.exe or inx-desktop.exe at InstallRoot" >&2
	exit 1
fi

# Inx.exe is the portable alias of the thin launcher.
cmp -s "$staging/Inx.exe" "$staging/inx-launcher.exe" || {
	echo "Inx.exe is not the packaged GUI launcher" >&2
	exit 1
}
if cmp -s "$staging/Inx.exe" "$staging/inx-cli.exe"; then
	echo "Inx.exe was overwritten by the CLI sidecar" >&2
	exit 1
fi

# Case-insensitive collision check among root exes.
actual=()
for path in "$staging"/*.exe; do
	[ -f "$path" ] || continue
	name="${path##*/}"
	folded=$(printf '%s' "$name" | tr '[:upper:]' '[:lower:]')
	if [ "${#actual[@]}" -gt 0 ]; then
		for previous in "${actual[@]}"; do
			previous_folded=$(printf '%s' "$previous" | tr '[:upper:]' '[:lower:]')
			if [ "$folded" = "$previous_folded" ]; then
				echo "Windows portable names collide case-insensitively: $previous and $name" >&2
				exit 1
			fi
		done
	fi
	actual+=("$name")
done
