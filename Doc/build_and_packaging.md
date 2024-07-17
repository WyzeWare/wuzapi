# Wuzapi Build and Packaging Documentation

This document outlines the process for building and packaging the Wuzapi application for multiple platforms.

## Required Tools

Before you can use the build script, you need to install the following tools:

1. **Go**: Used for compiling the application.
   - Installation: https://golang.org/doc/install

2. **dpkg-deb**: Used for creating Debian packages.
   - Installation on Debian/Ubuntu: `sudo apt-get install dpkg-dev`

3. **rpmbuild**: Used for creating RPM packages.
   - Installation on Fedora/CentOS: `sudo dnf install rpm-build`
   - Installation on Ubuntu: `sudo apt-get install rpm`

4. **tar**: Used for creating tarballs (usually pre-installed on most Unix-like systems).

5. **zip**: Used for creating ZIP files.
   - Installation on Debian/Ubuntu: `sudo apt-get install zip`
   - Installation on Fedora/CentOS: `sudo dnf install zip`

## Automatic installation

1. To use this script:
   - Make it executable: `chmod +x install_requirements.sh`
   - Run it: `sudo ./install_requirements.sh`

## Build Script

The `build_and_package.sh` script automates the process of building the application and creating packages for various platforms.

### Usage

1. Make the script executable:
 
```bash
chmod +x build_and_package.sh
```

2. Run the script:
```bash
./build_and_package.sh
```
### What the Script Does

1. Builds the Wuzapi application for Linux, macOS, and Windows.
2. Creates a Debian package (.deb) for Debian-based systems.
3. Creates an RPM package for Red Hat-based systems.
4. Creates a tarball for general Linux/Unix systems.
5. Creates a ZIP file for macOS.
6. Creates a ZIP file for Windows.

### Output

After running the script, you'll find the following files in your project directory:

- `wuzapi-linux-amd64`: Linux executable
- `wuzapi-darwin-amd64`: macOS executable
- `wuzapi-windows-amd64.exe`: Windows executable
- `wuzapi_1.0_amd64.deb`: Debian package
- `rpmbuild/RPMS/x86_64/wuzapi-1.0-1.x86_64.rpm`: RPM package
- `wuzapi-1.0-linux-amd64.tar.gz`: Linux tarball
- `wuzapi-1.0-darwin-amd64.zip`: macOS ZIP file
- `wuzapi-1.0-windows-amd64.zip`: Windows ZIP file

## Customization

You can customize the build process by modifying the following variables at the beginning of the `build_and_package.sh` script:

- `APP_NAME`: The name of your application
- `VERSION`: The version number of your application
- `MAINTAINER`: Your name and email address
- `DESCRIPTION`: A short description of your application

## Troubleshooting

If you encounter any issues:

1. Ensure all required tools are installed and accessible from the command line.
2. Check that you have the necessary permissions to create files and directories in your project folder.
3. If a specific packaging step fails, you can comment out the corresponding function call at the end of the script and re-run it.

## Additional Notes

- The script assumes your Go code is in the current directory and that `go build` can be run without additional flags or environment variables.
- For production use, you may want to add code signing for the Windows executable and notarization for the macOS application.
- Consider using a CI/CD pipeline for automated builds and testing before packaging.
