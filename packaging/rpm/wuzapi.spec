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
mkdir -p $RPM_BUILD_ROOT/%{_var}/log/wuzapi
echo "Log file created on %{DATE}" > $RPM_BUILD_ROOT/%{_var}/log/wuzapi/%{name}.log
chmod 755 $RPM_BUILD_ROOT/%{_bindir}/%{name}
chmod 755 $RPM_BUILD_ROOT/%{_sysconfdir}/wuzapi/install.sh
chmod 755 $RPM_BUILD_ROOT/%{_var}/log/wuzapi
chmod 644 $RPM_BUILD_ROOT/%{_var}/log/wuzapi/%{name}.log

%post
/bin/bash %{_sysconfdir}/wuzapi/install.sh
chown -R wuzapi:wuzapi %{_var}/log/wuzapi

%files
%attr(755, root, root) %{_bindir}/%{name}
%attr(755, root, root) %{_sysconfdir}/wuzapi/install.sh
%attr(755, wuzapi, wuzapi) %dir %{_var}/log/wuzapi
%attr(644, wuzapi, wuzapi) %{_var}/log/wuzapi/%{name}.log

%changelog
* Wed Jul 17 2024 Your Name <your.email@example.com> 1.0-1
- Initial RPM release