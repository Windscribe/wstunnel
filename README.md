# Windscribe Tunnel Proxy for client apps.
 This program forwards OpenVPN tcp traffic to WSTunnel or Stunnel server.

## Build
1. To build android library Run `build_android.sh`. (Require android sdk + ndk)
2. To build ios framework Run `build_ios.sh` (Requires xcode build tools)
3. To build binaries for desktop Run `build_desktop.sh`
4. Mobile can import it as libraries and desktop can use cli.


## Download from jitpack (Android only)
[![](https://jitpack.io/v/Windscribe/wstunnel.svg)](https://jitpack.io/#Windscribe/wstunnel)

## Dependencies
1. Gorrila web socket for wstunnel
   https://github.com/gorilla/websocket
2. Cobra for cli
   https://github.com/spf13/cobra
3. Zap for logging
   https://github.com/uber-go/zap
   
