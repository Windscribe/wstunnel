#!/usr/bin/env bash
set -e

rm -rf build
platforms=("windows/arm64" "windows/amd64" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

build_platform() {
    local goos=$1 goarch=$2
    local output_name="build/desktop/wstunnel-$goos-$goarch"
    [ "$goos" = "windows" ] && output_name+=".exe"

    mkdir -p build/desktop
    GOOS="$goos" GOARCH="$goarch" go build -o "$output_name" -a -gcflags=all="-l -B" -ldflags="-w -s"
    echo "Built $goos/$goarch at $output_name"
}

for platform in "${platforms[@]}"; do
    build_platform "${platform%%/*}" "${platform##*/}"
done
