#!/usr/bin/env sh
# install.sh — download and install wiki-mcp binary
# Usage: curl -sSfL https://github.com/robert2411/wiki-mcp/releases/latest/download/install.sh | sh
set -eu

REPO="robert2411/wiki-mcp"
BINARY="wiki-mcp"

# Resolve OS
OS="$(uname -s)"
case "${OS}" in
  Linux)  OS="Linux" ;;
  Darwin) OS="Darwin" ;;
  *)
    echo "Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

# Resolve arch
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64)    ARCH="x86_64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

# Detect latest version if not specified
if [ -z "${WIKI_MCP_VERSION:-}" ]; then
  WIKI_MCP_VERSION="$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  [ -n "${WIKI_MCP_VERSION}" ] || { echo "error: failed to detect latest version" >&2; exit 1; }
fi

# goreleaser strips the leading v from archive names (e.g. v1.2.3 → 1.2.3)
ARCHIVE="${BINARY}_${WIKI_MCP_VERSION#v}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${WIKI_MCP_VERSION}"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Downloading ${BINARY} ${WIKI_MCP_VERSION} (${OS}/${ARCH})…"
curl -sSfL "${BASE_URL}/${ARCHIVE}" -o "${TMP}/${ARCHIVE}"

# Verify checksum before extracting
curl -sSfL "${BASE_URL}/checksums.txt" -o "${TMP}/checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  (cd "${TMP}" && grep "${ARCHIVE}" checksums.txt | sha256sum -c -)
elif command -v shasum >/dev/null 2>&1; then
  (cd "${TMP}" && grep "${ARCHIVE}" checksums.txt | shasum -a 256 -c -)
else
  echo "Warning: no sha256sum or shasum found — skipping checksum verification" >&2
fi

tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"

# chmod before mv so the binary is never briefly executable at the destination
chmod +x "${TMP}/${BINARY}"

# Try /usr/local/bin first; fall back to $HOME/.local/bin
if mv "${TMP}/${BINARY}" "/usr/local/bin/${BINARY}" 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "${INSTALL_DIR}"
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
echo "Version: $("${INSTALL_DIR}/${BINARY}" --version)"

# Warn if install dir not in PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Warning: ${INSTALL_DIR} is not in PATH. Add it:" >&2
     echo "  export PATH=\"\${PATH}:${INSTALL_DIR}\"" >&2 ;;
esac
