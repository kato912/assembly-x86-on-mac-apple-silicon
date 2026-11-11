#!/usr/bin/env bash
# Turn the interesting lines of a failed `go test` run into a GitHub Actions
# error annotation, so failures are readable from the run summary (and the
# public API) without opening the full log.
#   scripts/annotate-failures.sh test-output.txt
f=${1:-test-output.txt}
[ -f "$f" ] || { echo "::error::no test output captured"; exit 0; }
lines=$(grep -E -A12 -- '--- FAIL|^FAIL|panic:|cannot|undefined|error' "$f" | grep -v '^=== RUN' | head -80)
[ -n "$lines" ] || lines=$(tail -40 "$f")
# Annotations are one line: encode %, CR and LF as GitHub expects.
msg=$(printf '%s' "$lines" | sed -e 's/%/%25/g' -e 's/\r/%0D/g' | awk 'BEGIN{ORS="%0A"} {print}')
echo "::error title=go test failed::$msg"
