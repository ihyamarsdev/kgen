#!/usr/bin/env bash

set -euo pipefail

OWNER="ihyamarsdev"
REPO="kgen"
DEFAULT_VERSION="v0.1.0"

# Fetch latest version tag from GitHub API
VERSION=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep -Po '"tag_name": "\K[^"]*' || echo "${DEFAULT_VERSION}")

# Detect OS
OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
case "${OS_TYPE}" in
  linux*)   OS="linux" ;;
  darwin*)  OS="darwin" ;;
  *)
    echo "Error: Unsupported OS type: ${OS_TYPE}"
    exit 1
    ;;
esac

# Detect Architecture
ARCH_TYPE=$(uname -m)
case "${ARCH_TYPE}" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture: ${ARCH_TYPE}"
    exit 1
    ;;
esac

# Formulate download filename and URL
BINARY_NAME="kgen-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

echo "Installing KGen ${VERSION} for ${OS}/${ARCH}..."
echo "Downloading from: ${DOWNLOAD_URL}"

# Temporary download file
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

curl -sSfL "${DOWNLOAD_URL}" -o "${TMP_DIR}/kgen"

chmod +x "${TMP_DIR}/kgen"

# Installation destination
INSTALL_DIR="/usr/local/bin"

if [ -w "${INSTALL_DIR}" ]; then
  mv "${TMP_DIR}/kgen" "${INSTALL_DIR}/kgen"
  echo "Successfully installed KGen to ${INSTALL_DIR}/kgen"
else
  echo "Need sudo privileges to write to ${INSTALL_DIR}"
  if command -v sudo >/dev/null 2>&1; then
    sudo mv "${TMP_DIR}/kgen" "${INSTALL_DIR}/kgen"
    echo "Successfully installed KGen to ${INSTALL_DIR}/kgen"
  else
    # Fallback to local user bin or current directory
    LOCAL_BIN="${HOME}/.local/bin"
    mkdir -p "${LOCAL_BIN}"
    mv "${TMP_DIR}/kgen" "${LOCAL_BIN}/kgen"
    echo "Installed locally to ${LOCAL_BIN}/kgen"
    echo "Make sure ${LOCAL_BIN} is in your PATH."
  fi
fi

echo "Installation complete. Verify with:"
echo "  kgen --help"
