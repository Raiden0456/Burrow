#!/bin/sh
set -e

for dir in /data "$CODEX_HOME"; do
	mkdir -p "$dir"
	if [ "$(stat -c %U "$dir")" != app ]; then
		chown -R app:app "$dir"
	fi
done

exec su-exec app "$@"
