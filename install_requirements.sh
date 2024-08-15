#!/bin/bash

# Function to install a package if it's not already installed
install_if_needed() {
    PACKAGE_NAME=$1
    INSTALL_COMMAND=$2

    if ! dpkg -l | grep -q $PACKAGE_NAME; then
        echo "Installing $PACKAGE_NAME..."
        eval $INSTALL_COMMAND
    else
        echo "$PACKAGE_NAME is already installed."
    fi
}

# Install Go
install_if_needed "golang-go" "sudo apt-get install -y golang-go"

# Install dpkg-dev (for dpkg-deb)
install_if_needed "dpkg-dev" "sudo apt-get install -y dpkg-dev"

# Install rpm (for rpmbuild)
install_if_needed "rpm" "sudo apt-get install -y rpm"

# Install tar
install_if_needed "tar" "sudo apt-get install -y tar"

# Install zip
install_if_needed "zip" "sudo apt-get install -y zip"

# Install OpenSSL development libraries
install_if_needed "libssl-dev" "sudo apt-get install -y libssl-dev"

# macOS-specific OpenSSL setup (optional)
if [[ "$OSTYPE" == "darwin"* ]]; then
    if ! brew ls --versions openssl >/dev/null; then
        echo "Installing OpenSSL via Homebrew..."
        brew install openssl
    else
        echo "OpenSSL is already installed via Homebrew."
    fi

    echo "Setting up environment variables for OpenSSL..."
    export PKG_CONFIG_PATH=$(brew --prefix openssl)/lib/pkgconfig
    export CGO_LDFLAGS="-L$(brew --prefix openssl)/lib"
    export CGO_CFLAGS="-I$(brew --prefix openssl)/include"
fi

echo "All required tools are installed and configured."
