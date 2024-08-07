#!/bin/sh

# Common variables
export GOOS=ios
export GOARCH=arm64
export CGO_ENABLED=1

build() {
    local sdk=$1
    local min_version=$2
    local platform_sdk=$3

    export SDK=$sdk
    CC="clang -arch arm64 -isysroot /Applications/Xcode.app/Contents/Developer/Platforms/${platform_sdk}.platform/Developer/SDKs/${platform_sdk}17.5.sdk -m${sdk}-version-min=${min_version} -fembed-bitcode"
    CGO_CFLAGS=""
    CGO_LDFLAGS="-framework CoreFoundation"
    export CC

    output_dir="./build/${sdk}/arm64"
    rm -rf "$output_dir"
    mkdir -p "$output_dir"
    go build -buildmode=c-archive -o "$output_dir/proxy.a" cli.go logger.go httpclient.go stunnelbidirection.go websocketbidirConnection.go common.go
}

# Build for Apple TVOS
build "appletvos" "17.0" "AppleTVOS"
echo "Apple TVOS framework at ./build/appletvos/arm64/proxy.a"

# Build for Apple TV Simulator
build "appletvsimulator" "17.0" "AppleTVSimulator"
echo "Apple TVOS framework at ./build/appletvsimulator/arm64/proxy.a"

# Build for iPhoneOS
build "iphoneos" "12.0" "iPhoneOS"
echo "iPhoneOS framework at ./build/iphoneos/arm64/proxy.a"

# Build for iPhoneSimulator
build "iphonesimulator" "12.0" "iPhoneSimulator"
echo "iPhoneSimulator framework at ./build/iphonesimulator/arm64/proxy.a"

# Create a combined headers directory
combined_headers_dir="./build/combined_headers"
rm -rf "$combined_headers_dir"
mkdir -p "$combined_headers_dir"

# Copy header files from all builds
cp ./build/appletvos/arm64/*.h "$combined_headers_dir"
# same for all platforms
#cp ./build/iphoneos/arm64/*.h "$combined_headers_dir"
#cp ./build/iphonesimulator/arm64/*.h "$combined_headers_dir"

# Create .xcframework
rm -rf ./build/Proxy.xcframework
xcodebuild -create-xcframework \
    -library ./build/appletvos/arm64/proxy.a -headers "$combined_headers_dir" \
    -library ./build/appletvsimulator/arm64/proxy.a -headers "$combined_headers_dir" \
    -library ./build/iphoneos/arm64/proxy.a -headers "$combined_headers_dir" \
    -library ./build/iphonesimulator/arm64/proxy.a -headers "$combined_headers_dir" \
    -output ./build/Proxy.xcframework

echo "Combined framework at ./build/Proxy.xcframework"