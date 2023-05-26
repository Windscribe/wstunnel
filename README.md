# Windscribe Tunnel Proxy for mobile apps.
 This library forwards OpenVPN tcp traffic to WSTunnel or Stunnel server.

## Build and use
1. Run `go mod tidy` in wstunnel directory.
2. Install gomobile tools if not already installed.
   [Download](https://github.com/golang/mobile).
3. To build android library Run `gomobile bind -o proxy.aar  -javapkg com.windscribe` in ./proxy.
4. To build ios framework Run `gomobile bind -target ios/arm64 -o proxy.xcframework` in ./proxy
   This builds platform specific libraries and bindings from exported functions.
5. Exported binding are used by the host app.
6. Import libraries in to project.


## Android using gradle
`allprojects {
repositories {
maven {
name "jitpack"
url "https://jitpack.io"
     } 
  }
}`

`implementation 'com.github.Ginder-Singh:wstest:1.0.0'`
