#!/bin/sh
export GOOS=ios
export GOARCH=arm64
export CGO_ENABLED=1
export SDK=iphoneos
export CGO_CFLAGS="-fembed-bitcode"
export MIN_VERSION=15
. ./target.sh
export CGO_LDFLAGS="-target ${TARGET} -syslibroot \"${SDK_PATH}\""
CC="$(pwd)/clangwrap.sh"
export CC
rm -r build/ios
output_dir="./build/ios/arm64"
mkdir -p build/ios/arm64
go build -buildmode=c-archive -o $output_dir/wstunnel.a cli.go logger.go httpclient.go stunnelbidirection.go websocketbidirConnection.go common.go