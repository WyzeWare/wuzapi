#!/bin/bash

# Set variables
APP_NAME="wuzapi"
VERSION="1.0"
MAINTAINER="Buenos Aires - Argentina"
DESCRIPTION="WuzApi clone is a Go-based application designed to facilitate webhook interactions. \
It provides functionality to handle regular messages and messages with file attachments \
by sending POST requests to specified URLs. The application includes features to update \
user information, log payloads, and manage webhook calls efficiently."

# Function to build for multiple platforms
build_multi_platform() {
    echo "Building for multiple platforms..."
    GOOS=linux GOARCH=amd64 go build -o ${APP_NAME}-linux-amd64 .
    GOOS=darwin GOARCH=amd64 go build -o ${APP_NAME}-darwin-amd64 .
    GOOS=windows GOARCH=amd64 go build -o ${APP_NAME}-windows-amd64.exe .
}

# Function to create Debian package
create_deb() {
    echo "Creating Debian package..."
    mkdir -p ${APP_NAME}_deb/DEBIAN
    mkdir -p ${APP_NAME}_deb/usr/local/bin

    cp ${APP_NAME}-linux-amd64 ${APP_NAME}_deb/usr/local/bin/${APP_NAME}

    cat << EOF > ${APP_NAME}_deb/DEBIAN/control
Package: $APP_NAME
Version: $VERSION
Section: custom
Priority: optional
Architecture: amd64
Essential: no
Installed-Size: 1024
Maintainer: $MAINTAINER
Description: $DESCRIPTION
EOF

    dpkg-deb --build ${APP_NAME}_deb
    mv ${APP_NAME}_deb.deb ${APP_NAME}_${VERSION}_amd64.deb

    echo "Debian package created: ${APP_NAME}_${VERSION}_amd64.deb"
    rm -rf ${APP_NAME}_deb
}

# Function to create RPM package
create_rpm() {
    echo "Creating RPM package..."
    mkdir -p rpmbuild/{SPECS,SOURCES,BUILD,RPMS,SRPMS}

    cat << EOF > rpmbuild/SPECS/${APP_NAME}.spec
Name:           $APP_NAME
Version:        $VERSION
Release:        1%{?dist}
Summary:        $DESCRIPTION
License:        MIT
URL:            https://github.com/WyzeWare/wuzapi
Source0:        %{name}-%{version}.tar.gz

%description
$DESCRIPTION

%prep

%install
rm -rf \$RPM_BUILD_ROOT
mkdir -p \$RPM_BUILD_ROOT/%{_bindir}
cp %{_sourcedir}/${APP_NAME}-linux-amd64 \$RPM_BUILD_ROOT/%{_bindir}/${APP_NAME}

%files
%{_bindir}/$APP_NAME

%changelog
* $(date +"%a %b %d %Y") $MAINTAINER $VERSION-1
- Initial RPM release
EOF

    cp ${APP_NAME}-linux-amd64 rpmbuild/SOURCES/
    
    echo "Building RPM package..."

    rpmbuild --define "_topdir $(pwd)/rpmbuild" -bb rpmbuild/SPECS/${APP_NAME}.spec

    echo "RPM package created in rpmbuild/RPMS/"
    rm -rf rpmbuild
}

# Function to create tarball for Linux/Unix
create_tarball() {
    echo "Creating tarball..."
    tar -czf ${APP_NAME}-${VERSION}-linux-amd64.tar.gz ${APP_NAME}-linux-amd64
    echo "Tarball created: ${APP_NAME}-${VERSION}-linux-amd64.tar.gz"
}

# Function to create ZIP for macOS
create_macos_zip() {
    echo "Creating macOS ZIP..."
    zip ${APP_NAME}-${VERSION}-darwin-amd64.zip ${APP_NAME}-darwin-amd64
    echo "macOS ZIP created: ${APP_NAME}-${VERSION}-darwin-amd64.zip"
}

# Function to create ZIP for Windows
create_windows_zip() {
    echo "Creating Windows ZIP..."
    zip ${APP_NAME}-${VERSION}-windows-amd64.zip ${APP_NAME}-windows-amd64.exe
    echo "Windows ZIP created: ${APP_NAME}-${VERSION}-windows-amd64.zip"
}

# Main execution
build_multi_platform
create_deb
create_rpm
create_tarball
create_macos_zip
create_windows_zip

echo "Build and packaging complete!"