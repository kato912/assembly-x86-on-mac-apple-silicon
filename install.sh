#!/bin/sh
# nasmx86 installer for macOS and Linux (Windows: download the .zip from the
# Releases page instead).
#
#   curl -fsSL https://raw.githubusercontent.com/kato912/assembly-x86-on-mac-apple-silicon/main/install.sh | sh
#
# Options (environment variables):
#   NASMX86_VERSION=v1.0.0     install a specific version (default: latest)
#   NASMX86_INSTALL_DIR=DIR    install into DIR (default: /usr/local/bin if
#                              writable, otherwise ~/.local/bin)
#   NASMX86_NO_NASM=1          do not try to install nasm
#   NASMX86_NO_MODIFY_PATH=1   do not edit shell startup files
set -eu

REPO="kato912/assembly-x86-on-mac-apple-silicon"
BASE_URL="${NASMX86_BASE_URL:-https://github.com/$REPO/releases/download}"

say()  { printf '%s\n' "$*"; }
warn() { printf 'nasmx86 installer: %s\n' "$*" >&2; }
die()  { warn "$*"; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

fetch() { # fetch URL FILE
    if have curl; then
        curl -fsSL --retry 3 -o "$2" "$1"
    elif have wget; then
        wget -q -O "$2" "$1"
    else
        die "need curl or wget to download files"
    fi
}

# --- platform -------------------------------------------------------------
case "$(uname -s)" in
    Darwin) os=darwin ;;
    Linux)  os=linux ;;
    MINGW*|MSYS*|CYGWIN*)
        die "on Windows, download nasmx86_..._windows_amd64.zip from https://github.com/$REPO/releases and see INSTALL-WINDOWS.txt" ;;
    *) die "unsupported operating system: $(uname -s)" ;;
esac
case "$(uname -m)" in
    x86_64|amd64)  arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) die "unsupported CPU: $(uname -m)" ;;
esac
# A shell running under Rosetta reports x86_64 on Apple Silicon: prefer the
# native arm64 build.
if [ "$os" = darwin ] && [ "$arch" = amd64 ] &&
   [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)" = 1 ]; then
    arch=arm64
fi

# --- version --------------------------------------------------------------
version="${NASMX86_VERSION:-latest}"
if [ "$version" = latest ]; then
    if have curl; then
        url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest") ||
            die "cannot reach GitHub to find the latest release"
    else
        url=$(wget -q -S --spider "https://github.com/$REPO/releases/latest" 2>&1 | sed -n 's/^ *Location: *//p' | tail -1)
    fi
    version="${url##*/}"
    case "$version" in
        v[0-9]*) ;;
        *) die "could not find the latest release (got '$version')" ;;
    esac
fi
case "$version" in v*) ;; *) version="v$version" ;; esac
num="${version#v}"
asset="nasmx86_${num}_${os}_${arch}.tar.gz"

# --- download and verify ----------------------------------------------------
tmp=$(mktemp -d 2>/dev/null || mktemp -d -t nasmx86)
trap 'rm -rf "$tmp"' EXIT INT TERM
say "Downloading nasmx86 $version for $os/$arch..."
fetch "$BASE_URL/$version/$asset" "$tmp/$asset" || die "download failed: $BASE_URL/$version/$asset"
if fetch "$BASE_URL/$version/checksums.txt" "$tmp/checksums.txt" 2>/dev/null; then
    want=$(grep " $asset\$" "$tmp/checksums.txt" | cut -d' ' -f1)
    if have sha256sum; then
        got=$(sha256sum "$tmp/$asset" | cut -d' ' -f1)
    elif have shasum; then
        got=$(shasum -a 256 "$tmp/$asset" | cut -d' ' -f1)
    else
        got="$want"
    fi
    [ -n "$want" ] && [ "$want" = "$got" ] || die "checksum mismatch for $asset"
fi
tar -xzf "$tmp/$asset" -C "$tmp"
[ -f "$tmp/nasmx86" ] || die "archive does not contain nasmx86"

# --- install ----------------------------------------------------------------
dir="${NASMX86_INSTALL_DIR:-}"
if [ -z "$dir" ]; then
    if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
        dir=/usr/local/bin
    else
        dir="$HOME/.local/bin"
    fi
fi
mkdir -p "$dir"
cp "$tmp/nasmx86" "$dir/nasmx86.new"
chmod 755 "$dir/nasmx86.new"
mv -f "$dir/nasmx86.new" "$dir/nasmx86"
# nasm-x86 and yasm-x86: the real nasm and yasm under names that say x86.
ln -sf nasmx86 "$dir/nasm-x86"
ln -sf nasmx86 "$dir/yasm-x86"
if [ "$os" = darwin ]; then
    # Files fetched by curl are not quarantined, but clear the flag in case a
    # browser-downloaded copy was involved, so Gatekeeper does not block it.
    xattr -d com.apple.quarantine "$dir/nasmx86" 2>/dev/null || true
fi
if [ -d "$tmp/examples" ]; then
    share="$HOME/.local/share/nasmx86"
    mkdir -p "$share" && cp -R "$tmp/examples" "$share/" 2>/dev/null || true
fi
say "Installed $dir/nasmx86"

# --- PATH -------------------------------------------------------------------
case ":$PATH:" in
    *":$dir:"*) on_path=1 ;;
    *) on_path=0 ;;
esac
if [ "$on_path" = 0 ] && [ "${NASMX86_NO_MODIFY_PATH:-0}" != 1 ]; then
    case "$(basename "${SHELL:-sh}")" in
        zsh)  rc="$HOME/.zshrc" ;;
        bash) if [ "$os" = darwin ]; then rc="$HOME/.bash_profile"; else rc="$HOME/.bashrc"; fi ;;
        fish) rc="" ;;
        *)    rc="$HOME/.profile" ;;
    esac
    line="export PATH=\"$dir:\$PATH\""
    if [ -n "$rc" ]; then
        if ! grep -qsF "$line" "$rc"; then
            printf '\n# added by the nasmx86 installer\n%s\n' "$line" >> "$rc"
            say "Added $dir to PATH in $rc (open a new terminal, or run: . \"$rc\")"
        fi
    else
        say "Add $dir to your PATH:  fish_add_path $dir"
    fi
fi

# --- nasm -------------------------------------------------------------------
if ! have nasm && [ "${NASMX86_NO_NASM:-0}" != 1 ]; then
    say "nasmx86 needs the nasm assembler; installing it..."
    sudo=""
    if [ "$(id -u)" != 0 ] && have sudo; then sudo=sudo; fi
    if [ "$os" = darwin ]; then
        if have brew; then
            brew install nasm || warn "brew install nasm failed"
        else
            warn "Homebrew not found. Install it from https://brew.sh, then run: brew install nasm"
        fi
    elif have apt-get; then $sudo apt-get install -y nasm || warn "could not install nasm (try: sudo apt install nasm)"
    elif have dnf;     then $sudo dnf install -y nasm || warn "could not install nasm (try: sudo dnf install nasm)"
    elif have yum;     then $sudo yum install -y nasm || warn "could not install nasm"
    elif have pacman;  then $sudo pacman -S --noconfirm nasm || warn "could not install nasm (try: sudo pacman -S nasm)"
    elif have zypper;  then $sudo zypper install -y nasm || warn "could not install nasm"
    elif have apk;     then $sudo apk add nasm || warn "could not install nasm"
    elif have brew;    then brew install nasm || warn "brew install nasm failed"
    else warn "please install nasm with your package manager"
    fi
fi

say ""
"$dir/nasmx86" --version
say "Try it:  nasmx86 $HOME/.local/share/nasmx86/examples/hello32.asm"
