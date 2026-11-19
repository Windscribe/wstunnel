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

# For ARM64
output_dir="./build/android/arm64-v8a"
TOOLCHAIN=("$ANDROID_NDK/toolchains/llvm/prebuilt/$PLATFORM-x86_64/bin/aarch64-linux-android21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=arm64 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For ARMv7
output_dir="./build/android/armeabi-v7a"
TOOLCHAIN=("$ANDROID_NDK/toolchains/llvm/prebuilt/$PLATFORM-x86_64/bin/armv7a-linux-androideabi21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=arm CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For x86
output_dir="./build/android/x86"
TOOLCHAIN=("$ANDROID_NDK/toolchains/llvm/prebuilt/$PLATFORM-x86_64/bin/i686-linux-android21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=386 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For x86_64
output_dir="./build/android/x86_64"
TOOLCHAIN=("$ANDROID_NDK/toolchains/llvm/prebuilt/$PLATFORM-x86_64/bin/x86_64-linux-android21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=amd64 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

echo 'Build successful...'