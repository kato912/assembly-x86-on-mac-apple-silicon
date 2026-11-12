#!/usr/bin/env bash
# Install gcc's 32-bit support (gcc -m32 and the i386 glibc) on a GitHub
# Actions Ubuntu runner, so the differential tests can link 32-bit
# programs that call the C library. apt is retried because the runners'
# mirrors sometimes fail.
set -u
if echo 'int main(void){return 0;}' | gcc -m32 -x c - -o /dev/null 2>/dev/null; then
    echo "gcc -m32 already works"; exit 0
fi
for attempt in 1 2 3; do
    if sudo apt-get install -y gcc-multilib ||
       { sudo apt-get update && sudo apt-get install -y gcc-multilib; }; then
        exit 0
    fi
    echo "::warning::apt-get install gcc-multilib failed (attempt $attempt)"
    sleep $((attempt * 10))
done
exit 1
