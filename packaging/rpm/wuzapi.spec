Name: %{APP_NAME}
Version: %{VERSION}
Release: 1%{?dist}
Summary: %{DESCRIPTION}
License: MIT
URL: https://github.com/WyzeWare/wuzapi
Source0: %{name}-%{version}.tar.gz
Source1: postinst
Source2: preinst
BuildArch: x86_64
Requires: openssl, sqlite

%description
%{DESCRIPTION}

%pre
/bin/bash %{Source2}

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir}
cp %{_sourcedir}/%{name}-linux-amd64 $RPM_BUILD_ROOT/%{_bindir}/%{name}
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/wuzapi
cp %{SOURCE1} $RPM_BUILD_ROOT/%{_sysconfdir}/wuzapi/install.sh
chmod 755 $RPM_BUILD_ROOT/%{_bindir}/%{name}
chmod 755 $RPM_BUILD_ROOT/%{_sysconfdir}/wuzapi/install.sh

%post
/bin/bash %{_sysconfdir}/wuzapi/install.sh

%files
%attr(755, root, root) %{_bindir}/%{name}
%attr(755, root, root) %{_sysconfdir}/wuzapi/install.sh

%changelog
* Wed Jul 17 2024 Your Name <your.email@example.com> 1.0-1
- Initial RPM release