#!/usr/bin/env bash
set -e

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
    local sdk=$1 min_version=$2
    local sdk_path
    sdk_path=$(xcrun --sdk "$sdk" --show-sdk-path)

    export SDK=$sdk
    export CC="clang -arch arm64 -isysroot ${sdk_path} -m${sdk}-version-min=${min_version} -fembed-bitcode"
    export CGO_CFLAGS=""
    export CGO_LDFLAGS="-framework CoreFoundation"

    local output_dir="./build/${sdk}/arm64"
    rm -rf "$output_dir"
    mkdir -p "$output_dir"
    go build -buildmode=c-archive -o "$output_dir/proxy.a" cli.go
    echo "Built $sdk at $output_dir/proxy.a"
}

build "appletvos" "17.0"
build "appletvsimulator" "17.0"
build "iphoneos" "12.0"
build "iphonesimulator" "12.0"

combined_headers_dir="./build/combined_headers"
rm -rf "$combined_headers_dir"
mkdir -p "$combined_headers_dir"
cp ./build/appletvos/arm64/*.h "$combined_headers_dir"

rm -rf ./build/Proxy.xcframework
xcodebuild -create-xcframework \
    -library ./build/appletvos/arm64/proxy.a -headers "$combined_headers_dir" \
    -library ./build/appletvsimulator/arm64/proxy.a -headers "$combined_headers_dir" \
    -library ./build/iphoneos/arm64/proxy.a -headers "$combined_headers_dir" \
    -library ./build/iphonesimulator/arm64/proxy.a -headers "$combined_headers_dir" \
    -output ./build/Proxy.xcframework

echo "Combined framework at ./build/Proxy.xcframework"
