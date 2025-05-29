#!/usr/bin/env bash

set -e

OPENSHIELD_DIR="/etc/openshield"
CERT_DIR="$OPENSHIELD_DIR/certs"
AGENT_BIN="openshield-agent"
SYSTEMD_UNIT="openshield-agent.service"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_DIR="/etc/systemd/system"

# Create necessary directories if they don't exist
mkdir -p "$OPENSHIELD_DIR"

# Detect OS
OS="$(uname | tr '[:upper:]' '[:lower:]')"

echo "Detected OS: $OS"

# Require VERSION and MANAGER arguments
if [[ -z "$1" || -z "$2" ]]; then
    echo "Usage: $0 <VERSION|latest> <MANAGER_ADDRESS>"
    echo "Example: $0 latest 192.168.10.11"
    exit 1
fi
VERSION="$1"
MANAGER="$2"

# If version is "latest", fetch the latest release tag from GitHub API
if [[ "$VERSION" == "latest" ]]; then
    echo "Fetching latest version from GitHub..."
    VERSION=$(curl -s https://api.github.com/repos/YordanOpenShield/openshield-agent/releases/latest | grep -oP '"tag_name":\s*"\K(.*)(?=")')
    if [[ -z "$VERSION" ]]; then
        echo "Failed to fetch latest version."
        exit 1
    fi
    echo "Latest version is $VERSION"
fi

if [[ "$OS" == "linux" ]]; then
    # Download agent binary if not present
    if [[ ! -f "$AGENT_BIN" ]]; then
        AGENT_URL="https://github.com/YordanOpenShield/openshield-agent/releases/download/${VERSION}/openshield-agent-linux-amd64-${VERSION}"
        echo "Downloading agent binary from $AGENT_URL ..."
        curl -L -o "$AGENT_BIN" "$AGENT_URL"
        chmod +x "$AGENT_BIN"
    fi

    echo "Copying $AGENT_BIN to $INSTALL_DIR (requires sudo)..."
    sudo cp "$AGENT_BIN" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$AGENT_BIN"

    # Download systemd unit file if not present
    if [[ ! -f "$SYSTEMD_UNIT" ]]; then
        UNIT_URL="https://raw.githubusercontent.com/YordanOpenShield/openshield-agent/refs/heads/main/helpers/openshield-agent.service"
        echo "Downloading systemd unit file from $UNIT_URL ..."
        curl -L -o "$SYSTEMD_UNIT" "$UNIT_URL"
    fi

    # Replace <manager-address> in the systemd unit file with the actual manager address
    sed -i "s|<manager-address>|$MANAGER|g" "$SYSTEMD_UNIT"

    # Copy systemd unit file
    echo "Copying $SYSTEMD_UNIT to $SYSTEMD_DIR (requires sudo)..."
    sudo cp "$SYSTEMD_UNIT" "$SYSTEMD_DIR/"
    sudo systemctl daemon-reload
    sudo systemctl enable openshield-agent
    sudo systemctl start openshield-agent
    echo "OpenShield Agent installed and started as a systemd service."

elif [[ "$OS" == "darwin" ]]; then
    echo "Detected macOS. Installing agent binary..."
    if [[ ! -f "$AGENT_BIN" ]]; then
        AGENT_URL="https://github.com/YordanOpenShield/openshield-agent/releases/download/${VERSION}/openshield-agent-darwin-amd64-${VERSION}"
        echo "Downloading agent binary from $AGENT_URL ..."
        curl -L -o "$AGENT_BIN" "$AGENT_URL"
        chmod +x "$AGENT_BIN"
    fi
    cp "$AGENT_BIN" /usr/local/bin/
    chmod +x /usr/local/bin/$AGENT_BIN
    echo "Agent installed to /usr/local/bin."
    echo "Note: macOS uses launchd, not systemd. Please create a launchd plist if you want to run as a service."

else
    echo "Unsupported OS: $OS"
    exit 1
fi