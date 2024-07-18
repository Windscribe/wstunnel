#!/bin/sh

# Goos for both tv and iphoneos is "ios"
export GOOS=ios
export GOARCH=arm64
export CGO_ENABLED=1

# Build for Apple TVOS
export SDK=appletvos
CC="clang -arch arm64 -isysroot /Applications/Xcode.app/Contents/Developer/Platforms/AppleTVOS.platform/Developer/SDKs/AppleTVOS17.5.sdk -mtvos-version-min=17.0 -fembed-bitcode"
CGO_CFLAGS=""
CGO_LDFLAGS="-framework CoreFoundation"
export CC

rm -rf build/"$SDK"
output_dir="./build/"$SDK"/arm64"
mkdir -p "$output_dir"
go build -buildmode=c-archive -o "$output_dir/proxy_tv.a" cli.go logger.go httpclient.go stunnelbidirection.go websocketbidirConnection.go common.go

# Build for iPhoneOS
export SDK=iphoneos
CC="clang -arch arm64 -isysroot /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS17.5.sdk -miphoneos-version-min=12.0 -fembed-bitcode"
CGO_CFLAGS=""
CGO_LDFLAGS="-framework CoreFoundation"
export CC

rm -rf build/"$SDK"
output_dir="./build/"$SDK"/arm64"
mkdir -p "$output_dir"
go build -buildmode=c-archive -o "$output_dir/proxy.a" cli.go logger.go httpclient.go stunnelbidirection.go websocketbidirConnection.go common.go
echo "Apple TVOS framework at $output_dir/proxy_tv.a"
echo "iPhoneOS framework at $output_dir/proxy.a"