#!/usr/bin/env bash

# Formating the architecture name to adapt to Go style
go_style_arch_name() {
    local arch=$1
    case ${arch} in
    x86_64|amd64)
        echo "amd64"
        ;;
    aarch64|arm64)
        echo "arm64"
        ;;
    *)
        echo "ERROR: invalid Arch, ${arch}"
        exit 1
        ;;
    esac
}

# Formating the architecture name to adapt to Linux style
linux_style_arch_name() {
    local arch=$1
    case ${arch} in
    x86_64|amd64)
        echo "x86_64"
        ;;
    aarch64|arm64)
        echo "aarch64"
        ;;
    *)
        echo "ERROR: invalid Arch, ${arch}"
        exit 1
        ;;
    esac
}

go_style_local_arch() {
   local arch=$(uname -m)
   go_style_arch_name $arch
}
