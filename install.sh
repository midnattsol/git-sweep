#!/usr/bin/env bash
#
# git-sweep installer
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/midnattsol/git-sweep/develop/install.sh | bash
#
# Environment variables:
#   INSTALL_DIR - Installation directory (default: ~/.local/bin)
#   VERSION     - Specific version to install (default: latest)

set -euo pipefail

REPO="midnattsol/git-sweep"
BINARY_NAME="git-sweep"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${CYAN}==>${NC} $1"; }
success() { echo -e "${GREEN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}==>${NC} $1"; }
error() { echo -e "${RED}==>${NC} $1" >&2; exit 1; }

# Detect OS
detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       error "Unsupported OS: $(uname -s)" ;;
  esac
}

# Detect architecture
detect_arch() {
  case "$(uname -m)" in
    x86_64)  echo "amd64" ;;
    amd64)   echo "amd64" ;;
    arm64)   echo "arm64" ;;
    aarch64) echo "arm64" ;;
    *)       error "Unsupported architecture: $(uname -m)" ;;
  esac
}

# Get latest version from GitHub
get_latest_version() {
  curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
  info "Installing git-sweep..."

  OS=$(detect_os)
  ARCH=$(detect_arch)
  
  info "Detected: ${OS}/${ARCH}"

  # Get version
  if [[ -z "${VERSION:-}" ]]; then
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
  fi

  if [[ -z "$VERSION" ]]; then
    error "Could not determine version to install"
  fi

  info "Version: ${VERSION}"

  # Build download URL
  ARCHIVE_NAME="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

  # Create temp directory
  TMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TMP_DIR"' EXIT

  # Download
  info "Downloading ${ARCHIVE_NAME}..."
  if ! curl -sSL -o "${TMP_DIR}/${ARCHIVE_NAME}" "$DOWNLOAD_URL"; then
    error "Failed to download from ${DOWNLOAD_URL}"
  fi

  # Extract
  info "Extracting..."
  tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "$TMP_DIR"

  # Install
  mkdir -p "$INSTALL_DIR"
  mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  success "Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"

  # Create git alias
  info "Creating git alias..."
  git config --global alias.sweep '!git-sweep'
  success "Created alias: git sweep"

  # Check if INSTALL_DIR is in PATH
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    warn "${INSTALL_DIR} is not in your PATH"
    echo ""
    echo "Add this to your shell config (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
  fi

  echo ""
  success "Installation complete!"
  echo ""
  echo "Usage:"
  echo "  git sweep           # Safe mode (dry-run)"
  echo "  git sweep --execute # Actually delete branches"
  echo "  git sweep --nuke    # Interactive mode"
  echo ""
}

main "$@"
