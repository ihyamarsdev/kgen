#!/usr/bin/env bash

set -euo pipefail

OWNER="ihyamarsdev"
REPO="kgen"
DEFAULT_VERSION="v0.4.0"

# Fetch latest version tag from GitHub API
VERSION=$(curl -sf "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep -Po '"tag_name": "\K[^"]*' || true)

# Validate VERSION format — fallback or fail
if [[ ! "${VERSION}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Could not determine latest version. Using fallback ${DEFAULT_VERSION}."
    echo "Check your internet connection and GitHub API access."
    VERSION="${DEFAULT_VERSION}"
fi

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

    # Require a checksum tool — fail hard if neither is available.
    if command -v shasum >/dev/null 2>&1; then
        ACTUAL=$(shasum -a 256 "${TMP_DIR}/kgen" | awk '{print $1}')
    elif command -v sha256sum >/dev/null 2>&1; then
        ACTUAL=$(sha256sum "${TMP_DIR}/kgen" | awk '{print $1}')
    else
        echo "Error: Neither shasum nor sha256sum found. Cannot verify checksum."
        echo "Please install one of these tools or download manually from ${DOWNLOAD_URL}"
        exit 1
    fi

    if [ "${ACTUAL}" != "${EXPECTED}" ]; then
        echo "Error: Checksum mismatch!"
        echo "  Expected: ${EXPECTED}"
        echo "  Actual:   ${ACTUAL}"
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
        # Detect the current shell and add to the appropriate rc file.
        current_shell=$(basename "$SHELL")
        rc_file=""
        case "$current_shell" in
            bash) rc_file="${HOME}/.bashrc" ;;
            zsh)  rc_file="${HOME}/.zshrc" ;;
            *)    rc_file="${HOME}/.profile" ;;
        esac
        if [ -n "${rc_file}" ] && [ -f "${rc_file}" ]; then
            if ! grep -q "${LOCAL_BIN}" "${rc_file}" 2>/dev/null; then
                echo "export PATH=\"\$PATH:${LOCAL_BIN}\"" >> "${rc_file}"
                echo "Added ${LOCAL_BIN} to PATH in ${rc_file}"
            fi
        fi
    fi
fi

echo "Installation complete. Verify with:"
echo "  kgen --help"
