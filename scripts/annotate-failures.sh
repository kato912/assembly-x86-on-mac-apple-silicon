#!/usr/bin/env bash
# Turn the interesting lines of a failed `go test` run into GitHub Actions
# error annotations, so failures are readable from the run summary (and the
# public API) without opening the full log.
#   scripts/annotate-failures.sh test-output.txt
f=${1:-test-output.txt}
[ -f "$f" ] || { echo "::error::no test output captured"; exit 0; }
encode() { sed -e 's/%/%25/g' -e 's/\r/%0D/g' | awk 'BEGIN{ORS="%0A"} {print}'; }
# 1. which tests failed
fails=$(grep -E -- '--- FAIL|^FAIL|panic:' "$f" | head -40)
echo "::error title=failed tests::$(printf '%s' "${fails:-$(tail -20 "$f")}" | encode)"
# 2. the messages of each failure (t.Errorf output is indented file:line)
grep -E -A14 '^\s+[A-Za-z0-9_]+\.go:[0-9]+:' "$f" | head -150 > "$f.details"
if [ -s "$f.details" ]; then
    # split into chunks so no single annotation gets too long
    split -l 50 "$f.details" "$f.part."
    for p in "$f".part.*; do
        echo "::error title=failure details::$(encode < "$p")"
    done
fi
