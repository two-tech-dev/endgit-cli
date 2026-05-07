#!/bin/bash
# Endgit Installer for Linux

set -e

REPO="two-tech-dev/endgit"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="endgit"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}Endgit Installer${NC}"
echo -e "${CYAN}=================${NC}"
echo ""

if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    echo -e "${YELLOW}Running as root - installing system-wide to $INSTALL_DIR${NC}"
else
    echo -e "${YELLOW}Installing to user directory: $INSTALL_DIR${NC}"
fi

echo -e "${YELLOW}Fetching latest release information...${NC}"

if command -v curl &> /dev/null; then
    RESPONSE=$(curl -sSL "$API_URL" -H "User-Agent: endgit-installer")
elif command -v wget &> /dev/null; then
    RESPONSE=$(wget -qO- "$API_URL" --header="User-Agent: endgit-installer")
else
    echo -e "${RED}Error: curl or wget required${NC}"
    exit 1
fi

VERSION=$(echo "$RESPONSE" | grep -o '"tag_name": *"[^"]*"' | head -1 | sed 's/"tag_name": *"\(.*\)"/\1/')

if [ -z "$VERSION" ]; then
    echo -e "${RED}Failed to fetch version${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: $VERSION${NC}"

# detect platform
OS="linux"
ARCH="amd64"

if [[ "$(uname -s)" == "Darwin" ]]; then
    OS="darwin"
fi

ASSET_NAME="endgit-${OS}-${ARCH}"

DOWNLOAD_URL=$(echo "$RESPONSE" | grep -o "\"browser_download_url\": *\"[^\"]*${ASSET_NAME}\"" | head -1 | sed 's/"browser_download_url": *"\(.*\)"/\1/')

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}No matching binary found${NC}"
    echo "$RESPONSE" | grep -o '"name": *"[^"]*"' | sed 's/"name": *"\(.*\)"/  - \1/'
    exit 1
fi

FILE_NAME=$(basename "$DOWNLOAD_URL")

echo -e "${GREEN}Found: $FILE_NAME${NC}"
echo ""

mkdir -p "$INSTALL_DIR"

INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

echo -e "${YELLOW}Downloading...${NC}"

if command -v curl &> /dev/null; then
    curl -L "$DOWNLOAD_URL" -o "$INSTALL_PATH"
elif command -v wget &> /dev/null; then
    wget -O "$INSTALL_PATH" "$DOWNLOAD_URL"
fi

chmod +x "$INSTALL_PATH"

echo -e "${GREEN}Installed to: $INSTALL_PATH${NC}"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Add to PATH:${NC}"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo -e "${GREEN}Endgit installation complete${NC}"