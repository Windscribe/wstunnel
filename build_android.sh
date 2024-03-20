export PATH=$PATH:~/go/bin
go mod tidy
gomobile init
rm -r build
mkdir -p build/arm64-v8a
mkdir -p build/armeabi-v7a
mkdir -p build/x86
mkdir -p build/x86_64
export CGO_ENABLED=1
export CGO_CFLAGS="-fstack-protector-strong"
NDK="/Users/gindersingh/Downloads/android-ndk-r21e"
# shellcheck disable=SC2016
buildCommand='go build -ldflags "-s -w" -buildmode=c-shared -o "$output_dir/libproxy.so" cli.go logger.go httpclient.go stunnelbidirection.go websocketbidirConnection.go common.go'
echo "$buildCommand"

# For ARM64
output_dir="./build/arm64-v8a"
TOOLCHAIN=("$NDK/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android30-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=arm64 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For ARMv7
output_dir="./build/armeabi-v7a"
TOOLCHAIN=("$NDK/toolchains/llvm/prebuilt/darwin-x86_64/bin/armv7a-linux-androideabi30-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=arm CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For x86
output_dir="./build/x86"
TOOLCHAIN=("$NDK/toolchains/llvm/prebuilt/darwin-x86_64/bin/i686-linux-android30-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=386 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For x86_64
output_dir="./build/x86_64"
TOOLCHAIN=("$NDK/toolchains/llvm/prebuilt/darwin-x86_64/bin/x86_64-linux-android30-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=amd64 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

echo 'Build successful...'