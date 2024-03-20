module wstunnel

go 1.18

replace github.com/gorilla/websocket => ./websocket

require (
	github.com/gorilla/websocket v1.4.2
	github.com/refraction-networking/utls v1.3.2
	github.com/spf13/cobra v1.7.0
	go.uber.org/zap v1.23.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/gaukas/godicttls v0.0.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.15.15 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)
