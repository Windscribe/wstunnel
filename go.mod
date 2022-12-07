module wstunnel

go 1.18

replace github.com/gorilla/websocket => ./websocket

require (
	github.com/gorilla/websocket v1.4.2
	github.com/refraction-networking/utls v1.1.5
	go.uber.org/zap v1.23.0
	golang.org/x/mobile v0.0.0-20221019142327-406ed3a7b8e4
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
	golang.org/x/tools v0.1.12 // indirect
)
