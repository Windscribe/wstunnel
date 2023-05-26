# Windscribe Tunnel Proxy for mobile apps.
 This library forwards OpenVPN tcp traffic to WSTunnel or Stunnel server.

## Build
1. To build android library Run `build_android.sh`. (Require android sdk + ndk)
2. To build ios framework Run `build_ios.sh` (Requires xcode build tools)
3. This builds platform specific libraries and bindings from exported functions.
4. Import library/framework in to project.


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
