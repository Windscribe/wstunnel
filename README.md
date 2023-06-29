# Windscribe Tunnel Proxy for client apps.
 This program forwards OpenVPN tcp traffic to WSTunnel or Stunnel server.

## Build
1. To build android library Run `build_android.sh`. (Require android sdk + ndk)
2. To build ios framework Run `build_ios.sh` (Requires xcode build tools)
3. To build binaries for desktop Run `build_desktop.sh`


## Download from jitpack (Android only)
[![](https://jitpack.io/v/Windscribe/wstunnel.svg)](https://jitpack.io/#Windscribe/wstunnel)

## Use Library
Import Library/Framework & Start proxy.
```val logFile = File(appContext.filesDir, PROXY_LOG).path
    initialise(BuildConfig.DEV, logFile)
    registerTunnelCallback(callback)
    if (isWSTunnel) {
    val remote = "wss://$ip:$port/$PROXY_TUNNEL_PROTOCOL/$PROXY_TUNNEL_ADDRESS/$WS_TUNNEL_PORT"
    startProxy(":$PROXY_TUNNEL_PORT", remote, 1, mtu)
    } else {
    val remote = "https://$ip:$port"
    startProxy(":$PROXY_TUNNEL_PORT", remote, 2, mtu)
    }
```
## Start binary
```Flags:
-d, --dev                    Turns on verbose logging.
-h, --help                   help for root
-l, --listenAddress string   Local port for proxy > :65479 (default ":65479")
-f, --logFilePath string     Path to log file > file.log
-m, --mtu int                1500 (default 1500)
-r, --remoteAddress string   Wstunnel > wss://$ip:$port/tcp/127.0.0.1/$WS_TUNNEL_PORT  Stunnel > https://$ip:$port
-t, --tunnelType int         WStunnel > 1 , Stunnel > 2 (default 1)
$ cli -l :65479 -r wss://$ip:$port/tcp/127.0.0.1/$WS_TUNNEL_PORT -t 1 -m 1500 -f file.log -d true
$ cli -l :65479 -r https://$ip:$port -t 2 -m 1500 -f file.log -d true
```

## Dependencies
1. Gorrila web socket for wstunnel [Link](https://github.com/gorilla/websocket)
2. Cobra for cli [Link](https://github.com/spf13/cobra)
3. Zap for logging [Link](https://github.com/uber-go/zap)
