#!/bin/bash

# Set variables
APP_NAME="wuzapi"
VERSION="1.0"
MAINTAINER="WyzeWare. support@wyzepal.com"
DESCRIPTION="WuzApi clone is a Go-based application designed to facilitate webhook interactions. \
It provides functionality to handle regular messages and messages with file attachments \
by sending POST requests to specified URLs. The application includes features to update \
user information, log payloads, and manage webhook calls efficiently."

# Function to build for multiple platforms
build_multi_platform() {
    echo "Building for multiple platforms..."
    GOOS=linux GOARCH=amd64 go build -o ${APP_NAME}-linux-amd64 ./cmd/wuzapi
    GOOS=darwin GOARCH=amd64 go build -o ${APP_NAME}-darwin-amd64 ./cmd/wuzapi
    GOOS=windows GOARCH=amd64 go build -o ${APP_NAME}-windows-amd64.exe ./cmd/wuzapi
}

# Function to create Debian package
create_deb() {
    echo "Creating Debian package..."
    mkdir -p ${APP_NAME}_deb/DEBIAN
    mkdir -p ${APP_NAME}_deb/usr/local/bin
    mkdir -p ${APP_NAME}_deb/etc/wuzapi

    cp ${APP_NAME}-linux-amd64 ${APP_NAME}_deb/usr/local/bin/${APP_NAME}
    cp packaging/scripts/postinst ${APP_NAME}_deb/DEBIAN/postinst
    cp packaging/scripts/preinst ${APP_NAME}_deb/DEBIAN/preinst

    # Set correct permissions for postinst script
    chmod 0755 ${APP_NAME}_deb/DEBIAN/postinst
    chmod 0755 ${APP_NAME}_deb/DEBIAN/preinst

    # Create control file with variables replaced
    sed -e "s/\${VERSION}/$VERSION/" \
        -e "s/\${MAINTAINER}/$MAINTAINER/" \
        -e "s/\${DESCRIPTION}/$DESCRIPTION/" \
        packaging/debian/control > ${APP_NAME}_deb/DEBIAN/control

    dpkg-deb --build ${APP_NAME}_deb
    mv ${APP_NAME}_deb.deb ${APP_NAME}_${VERSION}_amd64.deb

    echo "Debian package created: ${APP_NAME}_${VERSION}_amd64.deb"
    rm -rf ${APP_NAME}_deb
}

# Function to create RPM package
create_rpm() {
    echo "Creating RPM package..."
    mkdir -p rpmbuild/{SPECS,SOURCES,BUILD,RPMS,SRPMS}

    cp packaging/scripts/postinst rpmbuild/SOURCES/
    cp packaging/rpm/wuzapi.spec rpmbuild/SPECS/

    cp ${APP_NAME}-linux-amd64 rpmbuild/SOURCES/
    
    echo "Building RPM package..."
    rpmbuild --define "_topdir $(pwd)/rpmbuild" \
             --define "APP_NAME ${APP_NAME}" \
             --define "VERSION ${VERSION}" \
             --define "DESCRIPTION ${DESCRIPTION}" \
             -bb rpmbuild/SPECS/wuzapi.spec

    echo "RPM package created in rpmbuild/RPMS/"
    rm -rf rpmbuild
}

# Function to create tarball for Linux/Unix
create_tarball() {
    echo "Creating tarball..."
    cp packaging/scripts/preinst ./preinstall.sh
    echo "Please run preinstall.sh before installing the application to ensure all dependencies are met." > README.txt
    tar -czf ${APP_NAME}-${VERSION}-linux-amd64.tar.gz ${APP_NAME}-linux-amd64 preinstall.sh README.txt
    echo "Tarball created: ${APP_NAME}-${VERSION}-linux-amd64.tar.gz"
}

# Function to create ZIP for macOS
create_macos_zip() {
    echo "Creating macOS ZIP..."
    cp packaging/scripts/preinst ./preinstall.sh
    echo "Please run preinstall.sh before installing the application to ensure all dependencies are met." > README.txt
    zip ${APP_NAME}-${VERSION}-darwin-amd64.zip ${APP_NAME}-darwin-amd64 preinstall.sh README.txt
    echo "macOS ZIP created: ${APP_NAME}-${VERSION}-darwin-amd64.zip"
}

# Function to create ZIP for Windows
create_windows_zip() {
    echo "Creating Windows ZIP..."
    cp packaging/scripts/preinst ./preinstall.sh
    echo "Please ensure all dependencies (OpenSSL, SQLite3, etc.) are installed before running the application. On Windows, you may need to install these manually." > README.txt
    zip ${APP_NAME}-${VERSION}-windows-amd64.zip ${APP_NAME}-windows-amd64.exe preinstall.sh README.txt
    echo "Windows ZIP created: ${APP_NAME}-${VERSION}-windows-amd64.zip"
}

# Main execution
build_multi_platform
create_deb
create_rpm
create_tarball
create_macos_zip
create_windows_zip