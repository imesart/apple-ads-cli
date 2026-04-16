#!/bin/sh
set -eu

REPO="${REPO:-imesart/apple-ads-cli}"
BINARY="${BINARY:-aads}"
VERSION="${VERSION:-latest}"

DEFAULT_INSTALL_DIR="/usr/local/bin"
if [ -n "${HOME:-}" ]; then
  DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
fi
INSTALL_DIR="${INSTALL_DIR:-${DEFAULT_INSTALL_DIR}}"

error() {
  echo "Error: $*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || error "required command not found: $1"
}

detect_os() {
  case "$(uname -s)" in
    Darwin)
      echo "darwin"
      ;;
    Linux)
      echo "linux"
      ;;
    *)
      error "unsupported OS: $(uname -s)"
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64)
      echo "amd64"
      ;;
    arm64 | aarch64)
      echo "arm64"
      ;;
    *)
      error "unsupported architecture: $(uname -m)"
      ;;
  esac
}

resolve_latest_version() {
  latest_url="https://github.com/${REPO}/releases/latest"
  effective_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$latest_url")"
  version="${effective_url##*/}"

  [ -n "$version" ] || error "could not resolve latest release version"
  [ "$version" != "latest" ] || error "could not resolve latest release version"

  echo "$version"
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi

  error "required command not found: sha256sum or shasum"
}

install_binary() {
  source_path="$1"
  target_path="${INSTALL_DIR}/${BINARY}"

  if mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ]; then
    if command -v install >/dev/null 2>&1; then
      install -m 755 "$source_path" "$target_path"
    else
      cp "$source_path" "$target_path"
      chmod 755 "$target_path"
    fi
    return
  fi

  command -v sudo >/dev/null 2>&1 || error "cannot write to $INSTALL_DIR and sudo is not available"

  echo "Installing to $INSTALL_DIR requires elevated permissions..."
  sudo mkdir -p "$INSTALL_DIR"
  if command -v install >/dev/null 2>&1; then
    sudo install -m 755 "$source_path" "$target_path"
  else
    sudo cp "$source_path" "$target_path"
    sudo chmod 755 "$target_path"
  fi
}

need awk
need chmod
need cp
need curl
need mkdir
need mktemp
need tar
need uname

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ "$VERSION" = "latest" ]; then
  VERSION="$(resolve_latest_version)"
fi

ASSET="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
CHECKSUMS="${BINARY}_${VERSION}_checksums.txt"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT INT TERM

echo "Downloading ${ASSET}..."
curl -fsSLo "${TMPDIR}/${ASSET}" "${BASE_URL}/${ASSET}"
curl -fsSLo "${TMPDIR}/${CHECKSUMS}" "${BASE_URL}/${CHECKSUMS}"

EXPECTED_SHA256="$(awk -v file="$ASSET" '$2 == file { print $1 }' "${TMPDIR}/${CHECKSUMS}")"
[ -n "$EXPECTED_SHA256" ] || error "checksum not found for $ASSET"

ACTUAL_SHA256="$(sha256_file "${TMPDIR}/${ASSET}")"
[ "$EXPECTED_SHA256" = "$ACTUAL_SHA256" ] || error "checksum mismatch for $ASSET"

tar -xzf "${TMPDIR}/${ASSET}" -C "$TMPDIR"
[ -f "${TMPDIR}/${BINARY}" ] || error "archive did not contain ${BINARY}"

install_binary "${TMPDIR}/${BINARY}"

echo "Installed ${BINARY} ${VERSION} to ${INSTALL_DIR}/${BINARY}"
case ":${PATH:-}:" in
  *":${INSTALL_DIR}:"*)
    ;;
  *)
    echo "Note: ${INSTALL_DIR} is not on PATH."
    echo "Add it to your shell profile, e.g.:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
"${INSTALL_DIR}/${BINARY}" version
