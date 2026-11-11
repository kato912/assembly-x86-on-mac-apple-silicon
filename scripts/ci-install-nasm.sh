#!/usr/bin/env bash
# Install nasm on a GitHub Actions Linux runner. apt is tried first; if it
# fails (broken mirrors or third-party repositories make apt-get exit 100),
# fall back to the runner's preinstalled Homebrew, then to building the
# official release from source. apt's error is kept as an annotation.
set -u
NASM_SRC_VERSION=2.16.03
command -v nasm >/dev/null && { nasm -v; exit 0; }

log=$(mktemp)
annotate() { # title file
    msg=$(tail -20 "$2" | sed -e 's/%/%25/g' -e 's/\r/%0D/g' | awk 'BEGIN{ORS="%0A"} {print}')
    echo "::warning title=$1::$msg"
}

if sudo apt-get install -y nasm >"$log" 2>&1 ||
   { sudo apt-get update >>"$log" 2>&1; sudo apt-get install -y nasm >>"$log" 2>&1; }; then
    nasm -v; exit 0
fi
annotate "apt-get could not install nasm" "$log"

for brew in /home/linuxbrew/.linuxbrew/bin/brew "$(command -v brew 2>/dev/null)"; do
    if [ -n "$brew" ] && [ -x "$brew" ] && "$brew" install nasm >"$log" 2>&1; then
        dir=$("$brew" --prefix)/bin
        echo "$dir" >> "$GITHUB_PATH"
        "$dir/nasm" -v; exit 0
    fi
done
[ -s "$log" ] && annotate "brew could not install nasm" "$log"

# Last resort: build from source into /usr/local.
cd "$(mktemp -d)" &&
curl -fsSL "https://www.nasm.us/pub/nasm/releasebuilds/$NASM_SRC_VERSION/nasm-$NASM_SRC_VERSION.tar.xz" | tar -xJ &&
cd "nasm-$NASM_SRC_VERSION" && ./configure --prefix=/usr/local >"$log" 2>&1 && make -j"$(nproc)" >>"$log" 2>&1 &&
sudo make install >>"$log" 2>&1 && nasm -v && exit 0
annotate "building nasm from source failed" "$log"
exit 1
