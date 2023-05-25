# Windscribe Tunnel Proxy for mobile apps.
 This proxy forwards OpenVPN tcp traffic to WSTunnel or Stunnel servers.

## Build
1. Run `go mod tidy` in ws directory.
2. Install gomobile tools if not already installed.
   [Download](https://github.com/golang/mobile).
3. To build android library Run `gomobile bind -o proxy.aar  -javapkg com.windscribe`
4. To build ios framework Run `gomobile bind -target ios/arm64 -o proxy.xcframework`
   This builds platform specific libraries and bindings from exported functions.
5. Exported binding are used by the host app.