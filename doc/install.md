# Wuzapi

Wuzapi is a [brief description of your application].

## Table of Contents
1. [Installation](#installation)
   - [Debian/Ubuntu](#debianubuntu)
   - [Red Hat/CentOS](#red-hatcentos)
   - [macOS](#macos)
   - [Windows](#windows)
2. [Direct download instructions](#Direct-Download-Installation)
3. [Usage](#usage)
4. [Configuration](#configuration)
5. [Building from Source](#building-from-source)
6. [Contributing](#contributing)
7. [License](#license)

## Installation

### Debian/Ubuntu

1. Download the latest .deb package from the [releases page](https://github.com/yourusername/wuzapi/releases).
2. Install the package using dpkg: `sudo dpkg -i wuzapi_1.0_amd64.deb`
3. During installation, you'll be prompted to choose between PostgreSQL and SQLite3 as your database.

### Red Hat/CentOS

1. Download the latest .rpm package from the [releases page](https://github.com/yourusername/wuzapi/releases).
2. Install the package using rpm: `sudo rpm -i wuzapi-1.0-1.x86_64.rpm`
3. During installation, you'll be prompted to choose between PostgreSQL and SQLite3 as your database.

### macOS

1. Download the latest macOS ZIP file from the [releases page](https://github.com/yourusername/wuzapi/releases).
2. Extract the ZIP file.
3. Move the `wuzapi` executable to a directory in your PATH, for example:
    `sudo mv wuzapi /usr/local/bin/`
5. Follow the prompts to choose your database.

### Windows

1. Download the latest Windows ZIP file from the [releases page](https://github.com/yourusername/wuzapi/releases).
2. Extract the ZIP file to a location of your choice.
3. Open a Command Prompt as Administrator and navigate to the extraction directory.
4. Run the post-installation script:

`wuzapi.exe --post-install`

5. Follow the prompts to choose your database.

## Direct Download Installation

If you've downloaded the Wuzapi binary directly, follow these steps to complete the installation:

1. Move the Wuzapi binary to a directory in your system PATH. For example:

   `sudo mv wuzapi /usr/local/bin/`

2. Make the binary executable:
   `sudo chmod +x /usr/local/bin/wuzapi`

3. Download the post-installation script:
  `sudo curl -o /tmp/wuzapi_postinst https://raw.githubusercontent.com/WyzeWare/wuzapi/main/packaging/scripts/postinst`

4. Make the post-installation script executable:
  `sudo chmod +x /tmp/wuzapi_postinst`

5. Run the post-installation script:
   `sudo /tmp/wuzapi_postinst`

   - This script will prompt you to choose between PostgreSQL and SQLite3 as your database.

6. The script will create a configuration file at /etc/wuzapi/config. You can verify the installation by checking this file:
   `cat /etc/wuzapi/config`

7. Clean up the post-installation script:
   `sudo rm /tmp/wuzapi_postinst`

8. You can now run Wuzapi by typing wuzapi in your terminal.

Note: If you prefer to install Wuzapi in a different location, make sure to update your system's PATH or use the full path when running the application.

For any issues or further configuration, please refer to our full documentation or open an issue on our GitHub repository.

## Usage

[Provide instructions on how to use your application]

## Configuration

After installation, the database choice is stored in `/etc/wuzapi/config`. You can modify this file to change the database type.

[Add any other configuration details specific to your application]

## Building from Source

To build Wuzapi from source:

1. Ensure you have Go 1.16 or later installed.
2. Clone the repository:

`git clone https://github.com/yourusername/wuzapi.git`
`cd wuzapi`
3. Build the application: `go build ./cmd/wuzapi`
4. (Optional) To create packages for different systems, run:
 `./packaging/scripts/build_and_package.sh`
 ## Contributing

TODO:

## License

MIT License