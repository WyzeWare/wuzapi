#!/bin/bash

LOG_FILE="/tmp/wuzapi_preinstall.log"
exec > >(tee -a "$LOG_FILE") 2>&1
echo "Starting Wuzapi pre-installation script at $(date)"

echo "Note: This script requires sudo privileges for package installation."
echo "Note: This script checks for 'psql' (PostgreSQL client), but does not install the full PostgreSQL server."

set -e

# Check if the wuzapi user exists, create if it doesn't
if ! id wuzapi > /dev/null 2>&1; then
    adduser --system --group --no-create-home --home /nonexistent wuzapi
fi

# Function to prompt for yes/no with timeout
prompt_yes_no() {
    local timeout_duration=30
    while true; do
        read -t "$timeout_duration" -p "$1 (y/n): " yn
        if [ $? -eq 142 ]; then
            echo "Prompt timed out after $timeout_duration seconds. Exiting."
            exit 1
        fi
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
echo "Wuzapi pre-installation completed at $(date)"
echo "Log file is available at $LOG_FILE"
