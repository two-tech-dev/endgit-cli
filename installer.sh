#!/bin/bash
# EndGit Installer for Linux/macOS

set -e

INSTALL_DIR="${1:-$HOME/.local/bin}"
EXE_PATH="$INSTALL_DIR/endgit"
REPO="two-tech-dev/endgit-cli"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

# Colors
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
WHITE='\033[0;37m'
NC='\033[0m'

echo -e "${CYAN}EndGit Installer${NC}"
echo -e "${CYAN}================${NC}"
echo ""

echo -e "${YELLOW}Fetching latest release information...${NC}"

RESPONSE=$(curl -sf -H "User-Agent: endgit-installer" "$API_URL") || {
    echo -e "${RED}Error: Failed to fetch release information.${NC}"
    exit 1
}

VERSION=$(echo "$RESPONSE" | grep -o '"tag_name": *"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Failed to parse release information.${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: $VERSION${NC}"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             ARCH="amd64" ;;
esac

# Find matching asset
find_asset() {
    local pattern="$1"
    echo "$RESPONSE" \
        | grep -o '"browser_download_url": *"[^"]*"' \
        | grep -i "$pattern" \
        | head -1 \
        | grep -o 'https://[^"]*'
}

DOWNLOAD_URL=$(find_asset "endgit-${OS}-${ARCH}")
[ -z "$DOWNLOAD_URL" ] && DOWNLOAD_URL=$(find_asset "${OS}")
[ -z "$DOWNLOAD_URL" ] && DOWNLOAD_URL=$(find_asset "linux")

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}Error: No asset found for your platform (${OS}/${ARCH}).${NC}"
    echo -e "${YELLOW}Available assets:${NC}"
    echo "$RESPONSE" | grep -o '"name": *"[^"]*"' | cut -d'"' -f4 | while read -r name; do
        echo "  - $name"
    done
    exit 1
fi

FILE_NAME=$(basename "$DOWNLOAD_URL")
echo -e "${GREEN}Found asset: $FILE_NAME${NC}"
echo ""

mkdir -p "$INSTALL_DIR"

echo -e "${YELLOW}Downloading $FILE_NAME...${NC}"
curl -fL --progress-bar -o "$EXE_PATH" "$DOWNLOAD_URL" || {
    echo -e "${RED}Error: Download failed.${NC}"
    exit 1
}

if [ ! -f "$EXE_PATH" ]; then
    echo -e "${RED}Error: Download failed — file not found.${NC}"
    exit 1
fi

FILE_SIZE=$(wc -c < "$EXE_PATH")
if [ "$FILE_SIZE" -eq 0 ]; then
    echo -e "${RED}Error: Downloaded file is empty.${NC}"
    rm -f "$EXE_PATH"
    exit 1
fi

chmod +x "$EXE_PATH"

FILE_SIZE_KB=$(awk "BEGIN {printf \"%.2f\", $FILE_SIZE/1024}")
echo -e "${GREEN}Installed to: $EXE_PATH (${FILE_SIZE_KB} KB)${NC}"
echo ""

# Add to PATH in shell profile
add_to_path() {
    local profile="$1"
    if [ -f "$profile" ] && ! grep -q "$INSTALL_DIR" "$profile"; then
        echo "" >> "$profile"
        echo "# EndGit" >> "$profile"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$profile"
        echo -e "${GREEN}Added $INSTALL_DIR to PATH in $profile${NC}"
        return 0
    fi
    return 1
}

echo -e "${YELLOW}Adding to PATH...${NC}"

if echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo -e "${GREEN}Already in PATH${NC}"
else
    ADDED=false
    for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        add_to_path "$profile" && ADDED=true && break
    done
    if [ "$ADDED" = false ]; then
        echo -e "${YELLOW}Could not auto-update shell profile. Add this manually:${NC}"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
    echo -e "${YELLOW}Restart terminal or run: source ~/.bashrc (or ~/.zshrc)${NC}"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo -e "${CYAN}Usage:${NC}"
echo -e "${WHITE}  endgit search <plugin>${NC}"
echo -e "${WHITE}  endgit install <plugin>${NC}"
echo -e "${WHITE}  endgit init${NC}"