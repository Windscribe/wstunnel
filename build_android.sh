export PATH=$PATH:~/go/bin
go mod tidy
rm -r build/android
mkdir -p build/android/arm64-v8a
mkdir -p build/android/armeabi-v7a
mkdir -p build/android/x86
mkdir -p build/android/x86_64
export CGO_ENABLED=1
export CGO_CFLAGS="-fstack-protector-strong"
# shellcheck disable=SC2016
buildCommand='go build -ldflags "-s -w" -buildmode=c-shared -o "$output_dir/libproxy.so" cli.go logger.go httpclient.go stunnelbidirection.go websocketbidirConnection.go common.go'
echo "$buildCommand"

# For ARM64
output_dir="./build/android/arm64-v8a"
TOOLCHAIN=("$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=arm64 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For ARMv7
output_dir="./build/android/armeabi-v7a"
TOOLCHAIN=("$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/darwin-x86_64/bin/armv7a-linux-androideabi21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=arm CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For x86
output_dir="./build/android/x86"
TOOLCHAIN=("$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/darwin-x86_64/bin/i686-linux-android21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=386 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

## For x86_64
output_dir="./build/android/x86_64"
TOOLCHAIN=("$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/darwin-x86_64/bin/x86_64-linux-android21-clang")
# shellcheck disable=SC2086
GOOS=android GOARCH=amd64 CC="${TOOLCHAIN[0]}" output_dir="$output_dir" sh -c "$buildCommand"
rm $output_dir/libproxy.h

echo 'Build successful...'