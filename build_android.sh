#!/usr/bin/env bash
set -e

export PATH=$PATH:~/go/bin
go mod tidy
rm -rf build/android
mkdir -p build/android/arm64-v8a
mkdir -p build/android/armeabi-v7a
mkdir -p build/android/x86
mkdir -p build/android/x86_64
export CGO_ENABLED=1
export CGO_CFLAGS="-fstack-protector-strong"

if [[ "$(uname)" == "Darwin" ]]; then
    PLATFORM="darwin"
elif [[ "$(uname)" == "Linux" ]]; then
    PLATFORM="linux"
else
    PLATFORM="unknown"
fi

# `extras ndk` (the CI image's sdkmanager wrapper) installs the NDK but never sets ANDROID_NDK;
# it always lands under <sdk root>/ndk/<version>, so find it there if it isn't already set.
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
echo "$buildCommand"

build_arch() {
    local abi=$1 goarch=$2 clang=$3
    local output_dir="./build/android/$abi"
    local cc="$ANDROID_NDK/toolchains/llvm/prebuilt/$PLATFORM-x86_64/bin/$clang"

    echo "=== Building $abi (GOARCH=$goarch, CC=$cc) ==="
    if [ ! -x "$cc" ]; then
        echo "Toolchain not found or not executable: $cc"
        exit 1
    fi

    GOOS=android GOARCH="$goarch" CC="$cc" output_dir="$output_dir" sh -c "$buildCommand"

    echo "--- contents of $output_dir after build ---"
    ls -la "$output_dir"

    if [ ! -f "$output_dir/libproxy.h" ]; then
        echo "libproxy.h was not produced in $output_dir - build did not actually succeed."
        exit 1
    fi
    rm "$output_dir/libproxy.h"
}

build_arch "arm64-v8a" "arm64" "aarch64-linux-android21-clang"
build_arch "armeabi-v7a" "arm" "armv7a-linux-androideabi21-clang"
build_arch "x86" "386" "i686-linux-android21-clang"
build_arch "x86_64" "amd64" "x86_64-linux-android21-clang"

echo 'Build successful...'