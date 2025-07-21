#!/bin/bash

# Initialize local build flag
local_build=false

# Parse command line arguments
for arg in "$@"; do
    case $arg in
    --local)
        local_build=true
        ;;
    *)
        echo "Unknown parameter: $arg"
        exit 1
        ;;
    esac
done

arch=$(uname -m)
os=$(uname -s)

case $arch in
x86_64)
    go_arch="amd64"
    ;;
aarch64)
    go_arch="arm64"
    ;;
armv7l)
    go_arch="arm"
    ;;
i386 | i686)
    go_arch="386"
    ;;
*)
    echo "Unsupported architecture: $arch"
    exit 1
    ;;
esac

case $os in
Linux)
    go_os="linux"
    ;;
Darwin)
    go_os="darwin"
    ;;
FreeBSD)
    go_os="freebsd"
    ;;
Windows_NT)
    go_os="windows"
    ;;
*)
    echo "Unsupported operating system: $os"
    exit 1
    ;;
esac

echo "Detected OS: $os ($go_os), architecture: $arch ($go_arch)"

GOOS=$go_os GOARCH=$go_arch CGO_ENABLED=0 go build -ldflags="-w -s" -o timeoracle

# Move the executable only if not in local mode
if [ "$local_build" = false ]; then
    mv timeoracle ../benchmarks/cmd/bin/
    echo "Moved executable to ../benchmarks/cmd/bin/"
else
    echo "Executable retained in current directory (--local mode)"
fi

echo "Build complete for $go_os/$go_arch"
