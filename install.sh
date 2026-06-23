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
CHECKSUM_URL="${DOWNLOAD_URL}.sha256"

echo "Installing KGen ${VERSION} for ${OS}/${ARCH}..."
echo "Downloading from: ${DOWNLOAD_URL}"

# Temporary download file
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

curl -sSfL "${DOWNLOAD_URL}" -o "${TMP_DIR}/kgen"

# Verify checksum if available.
if curl -sSfL -o /dev/null -w "%{http_code}" "${CHECKSUM_URL}" | grep -q "200"; then
  curl -sSfL "${CHECKSUM_URL}" -o "${TMP_DIR}/${BINARY_NAME}.sha256"
  EXPECTED=$(awk '{print $1}' "${TMP_DIR}/${BINARY_NAME}.sha256")
  ACTUAL=$(shasum -a 256 "${TMP_DIR}/kgen" 2>/dev/null || sha256sum "${TMP_DIR}/kgen" 2>/dev/null || echo "")
  ACTUAL_HASH=$(echo "${ACTUAL}" | awk '{print $1}')

  if [ -n "${ACTUAL_HASH}" ] && [ "${ACTUAL_HASH}" != "${EXPECTED}" ]; then
    echo "Error: Checksum mismatch!"
    echo "  Expected: ${EXPECTED}"
    echo "  Actual:   ${ACTUAL_HASH}"
    echo "The downloaded binary may have been tampered with. Aborting installation."
    exit 1
  fi
  echo "Checksum verified ✓"
fi

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
    # Add to shell profile if not already present.
    for rc in ~/.bashrc ~/.zshrc ~/.profile; do
      if [ -f "${rc}" ] && ! grep -q "${LOCAL_BIN}" "${rc}" 2>/dev/null; then
        echo "export PATH=\"\$PATH:${LOCAL_BIN}\"" >> "${rc}"
        echo "Added ${LOCAL_BIN} to PATH in ${rc}"
      fi
    done
  fi
fi

echo "Installation complete. Verify with:"
echo "  kgen --help"
