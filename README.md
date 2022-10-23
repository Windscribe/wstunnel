# Windscribe WSTunnel for mobile apps.
This project is a fork of [https://github.com/trazfr/tcp-over-websocket](https://github.com/trazfr/tcp-over-websocket). Only relevant source are kept for easier maintenance.

## Build
1. Run `go mod tidy` in ws directory.
2. Install gomobile tools if not already installed.
   [Download](https://github.com/golang/mobile).
3. To build android library Run `gomobile bind -o wstunnel.aar  -javapkg com.windscribe.websockettunnel`
4. To build ios framework Run `gomobile bind -target ios/arm64 -o wstunnel.xcframework`
   This builds platform specific libraries and bindings from exported functions.
5. Exported binding are used by the host app.

## What to do next
1. Test , test and test more.
2. Implement Ws tunnel protocol for iOS.