#!/bin/bash

# Update package lists
echo "Updating package lists..."
sudo apt-get update

# Install Go
echo "Installing Go..."
sudo apt-get install -y golang-go

# Install dpkg-dev (for Debian packaging)
echo "Installing dpkg-dev..."
sudo apt-get install -y dpkg-dev

# Install rpm (for RPM packaging)
echo "Installing rpm..."
sudo apt-get install -y rpm

# Install zip
echo "Installing zip..."
sudo apt-get install -y zip

# Verify installations
echo "Verifying installations..."
go version
dpkg-deb --version
rpm --version
zip --version

echo "All requirements have been installed."