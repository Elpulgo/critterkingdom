#!/bin/sh
# Exports the Aseprite sources in assets-src/ into assets/, which the game
# embeds:
#
#   assets-src/<path>.aseprite  ->  assets/<path>.png and assets/<path>.json
#
# The PNG is a packed sprite sheet and the JSON lists its frames and tags.
# Tilesets under assets-src/tiles/ export as a plain PNG without JSON.
# Outputs that are newer than their source are skipped. Set ASEPRITE to the
# Aseprite program when it is not on the PATH.
set -eu

cd "$(dirname "$0")/.."

src=assets-src
out=assets

if [ ! -d "$src" ]; then
	echo "export-art: no $src/ folder yet, nothing to export"
	exit 0
fi

sources=$(find "$src" -type f \( -name '*.aseprite' -o -name '*.ase' \) | sort)
if [ -z "$sources" ]; then
	echo "export-art: no Aseprite files in $src/, nothing to export"
	exit 0
fi

if [ -z "${ASEPRITE:-}" ]; then
	ASEPRITE=aseprite
	if ! command -v aseprite >/dev/null 2>&1; then
		for app in \
			"/Applications/Aseprite.app/Contents/MacOS/aseprite" \
			"$HOME/Library/Application Support/Steam/steamapps/common/Aseprite/Aseprite.app/Contents/MacOS/aseprite"; do
			if [ -x "$app" ]; then
				ASEPRITE=$app
				break
			fi
		done
	fi
fi
if ! command -v "$ASEPRITE" >/dev/null 2>&1; then
	echo "export-art: cannot find Aseprite. Install it, or set ASEPRITE to its path." >&2
	exit 1
fi

# fresh succeeds when every output exists and is newer than the source.
fresh() {
	fresh_source=$1
	shift
	for fresh_output in "$@"; do
		if [ ! -f "$fresh_output" ] || [ ! "$fresh_output" -nt "$fresh_source" ]; then
			return 1
		fi
	done
	return 0
}

exported=0
skipped=0
while IFS= read -r file; do
	rel=${file#"$src"/}
	name=$out/${rel%.*}
	case $rel in
	tiles/*)
		if fresh "$file" "$name.png"; then
			skipped=$((skipped + 1))
			continue
		fi
		mkdir -p "$(dirname "$name")"
		"$ASEPRITE" -b "$file" --save-as "$name.png"
		;;
	*)
		if fresh "$file" "$name.png" "$name.json"; then
			skipped=$((skipped + 1))
			continue
		fi
		mkdir -p "$(dirname "$name")"
		"$ASEPRITE" -b "$file" --sheet "$name.png" --data "$name.json" \
			--format json-array --list-tags --sheet-type packed
		;;
	esac
	echo "export-art: $rel"
	exported=$((exported + 1))
done <<LIST
$sources
LIST

echo "export-art: $exported exported, $skipped already up to date"
