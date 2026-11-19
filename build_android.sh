#!/usr/bin/env bash
set -e

export PATH=$PATH:~/go/bin
go mod tidy
rm -rf build/android
mkdir -p build/android/arm64-v8a build/android/armeabi-v7a build/android/x86 build/android/x86_64
export CGO_ENABLED=1
export CGO_CFLAGS="-fstack-protector-strong"

# The exported C functions (and the header) need `import "C"` and runtime/cgo, commented out in
# cli.go for the desktop build. Turn them on for this build and put cli.go back afterwards.
cp cli.go cli.go.orig
trap 'mv cli.go.orig cli.go' EXIT
sed -i.bak -e 's|^	//"C"$|	"C"|' -e 's|^	//_ "runtime/cgo"$|	_ "runtime/cgo"|' cli.go
rm -f cli.go.bak
grep -q '^	"C"$' cli.go

if [[ "$(uname)" == "Darwin" ]]; then
    PLATFORM="darwin"
elif [[ "$(uname)" == "Linux" ]]; then
    PLATFORM="linux"
else
    PLATFORM="unknown"
fi

# extras ndk (the CI image's sdkmanager wrapper) never sets ANDROID_NDK; find it under the SDK root.
if [ -z "$ANDROID_NDK" ]; then
    SDK_ROOT="${ANDROID_SDK_ROOT:-$ANDROID_HOME}"
    ANDROID_NDK=$(find "$SDK_ROOT/ndk" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort -V | tail -1)
fi
if [ -z "$ANDROID_NDK" ] || [ ! -d "$ANDROID_NDK" ]; then
    echo "ANDROID_NDK not found. Set ANDROID_NDK or install the NDK via the Android SDK manager."
    exit 1
fi
echo "Using ANDROID_NDK=$ANDROID_NDK"

# shellcheck disable=SC2016
buildCommand='go build -ldflags "-s -w" -buildmode=c-shared -o "$output_dir/libproxy.so" cli.go'

build_arch() {
    local abi=$1 goarch=$2 clang=$3
    local output_dir="./build/android/$abi"
    local cc="$ANDROID_NDK/toolchains/llvm/prebuilt/$PLATFORM-x86_64/bin/$clang"

    if [ ! -x "$cc" ]; then
        echo "Toolchain not found: $cc"
        exit 1
    fi

    GOOS=android GOARCH="$goarch" CC="$cc" output_dir="$output_dir" sh -c "$buildCommand"

    if [ ! -f "$output_dir/libproxy.h" ]; then
        echo "libproxy.h was not produced in $output_dir"
        exit 1
    fi
    rm "$output_dir/libproxy.h"
    echo "Built $abi at $output_dir/libproxy.so"
}

build_arch "arm64-v8a" "arm64" "aarch64-linux-android21-clang"
build_arch "armeabi-v7a" "arm" "armv7a-linux-androideabi21-clang"
build_arch "x86" "386" "i686-linux-android21-clang"
build_arch "x86_64" "amd64" "x86_64-linux-android21-clang"
