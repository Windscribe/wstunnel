#!/bin/zsh

CLANG=$(xcrun --sdk "$SDK" --find clang)

exec "$CLANG" -target "$TARGET" -isysroot "$SDK_PATH" "$@"