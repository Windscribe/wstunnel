#!/bin/sh
set -e

# Common variables
export GOOS=ios
export GOARCH=arm64
export CGO_ENABLED=1

# The exported C functions (and the header) need `import "C"` and runtime/cgo, commented out in
# cli.go for the desktop build. Turn them on for this build and put cli.go back afterwards.
cp cli.go cli.go.orig
trap 'mv cli.go.orig cli.go' EXIT
sed -i.bak -e 's|^	//"C"$|	"C"|' -e 's|^	//_ "runtime/cgo"$|	_ "runtime/cgo"|' cli.go
rm -f cli.go.bak
grep -q '^	"C"$' cli.go

build() {
    local sdk=$1
    local min_version=$2

    export SDK=$sdk
    # The installed SDK, whatever Xcode version the runner has.
    local sdk_path=$(xcrun --sdk "$sdk" --show-sdk-path)
    CC="clang -arch arm64 -isysroot ${sdk_path} -m${sdk}-version-min=${min_version} -fembed-bitcode"
    CGO_CFLAGS=""
    CGO_LDFLAGS="-framework CoreFoundation"
    export CC

    output_dir="./build/${sdk}/arm64"
    rm -rf "$output_dir"
    mkdir -p "$output_dir"
    go build -buildmode=c-archive -o "$output_dir/proxy.a" cli.go
}

# Build for Apple TVOS
build "appletvos" "17.0"
echo "Apple TVOS framework at ./build/appletvos/arm64/proxy.a"

# Build for Apple TV Simulator
build "appletvsimulator" "17.0"
echo "Apple TVOS framework at ./build/appletvsimulator/arm64/proxy.a"

# Build for iPhoneOS
build "iphoneos" "12.0"
echo "iPhoneOS framework at ./build/iphoneos/arm64/proxy.a"

# Build for iPhoneSimulator
build "iphonesimulator" "12.0"
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