#!/usr/bin/env bash
set -euo pipefail

REPO="duguyue100/midnight-captain"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="mc"
LOCAL_BUILD=0

# Parse args
for arg in "$@"; do
  case "$arg" in
    --local-build) LOCAL_BUILD=1 ;;
    --help|-h)
      echo "Usage: install.sh [--local-build]"
      echo ""
      echo "  (no args)      Download latest release from GitHub and install to ~/.local/bin"
      echo "  --local-build  Build from source (must run from repo root), install to ~/.local/bin"
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      exit 1
      ;;
  esac
done

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

info()    { echo "[midnight-captain] $*"; }
success() { echo "[midnight-captain] ✓ $*"; }
warn()    { echo "[midnight-captain] ⚠ $*" >&2; }
die()     { echo "[midnight-captain] ✗ $*" >&2; exit 1; }

check_cmd() { command -v "$1" >/dev/null 2>&1; }

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    linux)  os="linux" ;;
    darwin) os="darwin" ;;
    *) die "Unsupported OS: $os. Only linux and darwin are supported." ;;
  esac

  case "$arch" in
    x86_64)          arch="amd64" ;;
    amd64)           arch="amd64" ;;
    aarch64|arm64)   arch="arm64" ;;
    *) die "Unsupported architecture: $arch. Only amd64 and arm64 are supported." ;;
  esac

  echo "${os}-${arch}"
}

ensure_install_dir() {
  mkdir -p "${INSTALL_DIR}"
}

check_path() {
  if echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    return
  fi
  warn "${INSTALL_DIR} is not in your PATH."
  warn "Add this line to your shell config (~/.bashrc, ~/.zshrc, etc.):"
  warn ""
  warn "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  warn ""
}

# ---------------------------------------------------------------------------
# Local build mode
# ---------------------------------------------------------------------------

local_build() {
  info "Local build mode."

  check_cmd go  || die "'go' not found. Install Go from https://go.dev/dl/"
  check_cmd make || die "'make' not found."

  # Must be run from repo root (where Makefile lives)
  if [[ ! -f "Makefile" ]] || ! grep -q "midnight-captain\|cmd/mc" Makefile go.mod 2>/dev/null; then
    die "Run this script from the midnight-captain repo root."
  fi

  info "Building..."
  make build

  local src="bin/${BINARY_NAME}"
  [[ -f "$src" ]] || die "Build succeeded but ${src} not found."

  ensure_install_dir
  cp "${src}" "${INSTALL_DIR}/${BINARY_NAME}"
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  success "Installed midnight-captain → ${INSTALL_DIR}/${BINARY_NAME}"
  check_path
}

# ---------------------------------------------------------------------------
# Download mode
# ---------------------------------------------------------------------------

download_release() {
  local platform
  platform="$(detect_platform)"
  info "Detected platform: ${platform}"

  # Fetch latest release tag from GitHub API
  check_cmd curl || die "'curl' not found."

  info "Fetching latest release info..."
  local api_url="https://api.github.com/repos/${REPO}/releases/latest"
  local release_json
  release_json="$(curl -fsSL "${api_url}")" \
    || die "Failed to fetch release info from ${api_url}. Check your internet connection."

  # Extract tag name (works without jq)
  local tag
  tag="$(echo "${release_json}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  [[ -n "$tag" ]] || die "Could not determine latest release tag."

  local binary="mc-${platform}"
  local download_url="https://github.com/${REPO}/releases/download/${tag}/${binary}"

  info "Downloading ${binary} (${tag})..."
  local tmp_file
  tmp_file="$(mktemp)"
  trap 'rm -f "${tmp_file}"' EXIT

  curl -fSL --progress-bar "${download_url}" -o "${tmp_file}" \
    || die "Download failed: ${download_url}"

  ensure_install_dir
  mv "${tmp_file}" "${INSTALL_DIR}/${BINARY_NAME}"
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  success "Installed midnight-captain ${tag} → ${INSTALL_DIR}/${BINARY_NAME}"
  check_path
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

if [[ "$LOCAL_BUILD" -eq 1 ]]; then
  local_build
else
  download_release
fi
