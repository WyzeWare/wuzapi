#!/bin/bash

# Function to prompt for yes/no
prompt_yes_no() {
    while true; do
        read -p "$1 (y/n): " yn
        case $yn in
            [Yy]*) return 0 ;;
            [Nn]*) return 1 ;;
            *) echo "Please answer yes or no." ;;
        esac
    done
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install a package
install_package() {
    local package=$1
    if command_exists apt-get; then
        sudo apt-get update && sudo apt-get install -y "$package"
    elif command_exists yum; then
        sudo yum install -y "$package"
    elif command_exists brew; then
        brew install "$package"
    else
        echo "Unsupported package manager. Please install $package manually."
        exit 1
    fi
}

# Array of required commands
REQUIRED_COMMANDS=(openssl sqlite3 sudo psql wget)

# Check for each required command
for cmd in "${REQUIRED_COMMANDS[@]}"; do
    if ! command_exists "$cmd"; then
        echo "$cmd is not installed."
        if prompt_yes_no "Do you want to install $cmd?"; then
            install_package "$cmd"
        else
            echo "$cmd is required for the script to run properly. Exiting."
            exit 1
        fi
    else
        echo "$cmd is already installed."
    fi
done

echo "All required commands are installed. You can now run the Wuzapi installation script."
