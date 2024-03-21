clangwrap.sh#!/bin/sh

SDK_PATH=$(xcrun --sdk "$SDK" --show-sdk-path)
export SDK_PATH

if [ "$GOARCH" = "amd64" ]; then
    CARCH="x86_64"
elif [ "$GOARCH" = "arm64" ]; then
    CARCH="arm64"
fi

if [ "$SDK" = "iphoneos" ]; then
  export TARGET="$CARCH-apple-ios$MIN_VERSION"
elif [ "$SDK" = "iphonesimulator" ]; then
  export TARGET="$CARCH-apple-ios$MIN_VERSION-simulator"
fi