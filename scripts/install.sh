#!/usr/bin/env bash
set -euo pipefail

REPO="${AGENT_EQUIP_REPO:-agent-equip/agent-equip}"
INSTALL_DIR="${AGENT_EQUIP_INSTALL_DIR:-/usr/local/bin}"
VERSION="${AGENT_EQUIP_VERSION:-latest}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1"; exit 1; }
}

need_cmd curl
need_cmd uname

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$os" in
  linux) os="linux" ;;
  darwin) os="darwin" ;;
  msys*|mingw*|cygwin*) os="windows" ;;
  *) echo "Unsupported OS: $os"; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported arch: $arch"; exit 1 ;;
esac

if [[ "$VERSION" == "latest" ]]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p' | head -n1)
  [[ -n "$VERSION" ]] || { echo "Could not resolve latest version"; exit 1; }
fi

asset="agent-equip_${VERSION#v}_${os}_${arch}"
url_base="https://github.com/${REPO}/releases/download/${VERSION}"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

if [[ "$os" == "windows" ]]; then
  need_cmd unzip
  archive="${asset}.zip"
  curl -fsSL "${url_base}/${archive}" -o "$tmp_dir/$archive"
  unzip -q "$tmp_dir/$archive" -d "$tmp_dir"
  bin_src="$tmp_dir/agent-equip.exe"
  bin_dst="$INSTALL_DIR/agent-equip.exe"
else
  need_cmd tar
  archive="${asset}.tar.gz"
  curl -fsSL "${url_base}/${archive}" -o "$tmp_dir/$archive"
  tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"
  bin_src="$tmp_dir/agent-equip"
  bin_dst="$INSTALL_DIR/agent-equip"
fi

mkdir -p "$INSTALL_DIR"
install -m 0755 "$bin_src" "$bin_dst"

echo "Installed: $bin_dst"
"$bin_dst" --help || true
