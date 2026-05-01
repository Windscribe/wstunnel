module github.com/Windscribe/wstunnel

go 1.24

replace github.com/gorilla/websocket => ./websocket

require (
	github.com/gorilla/websocket v1.4.2
	github.com/refraction-networking/utls v1.8.2
	github.com/spf13/cobra v1.7.0
	go.uber.org/zap v1.23.0
)

require (
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)
