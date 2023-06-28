export PATH=$PATH:~/go/bin
go mod tidy
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
mkdir build
cd proxy
gomobile bind -target ios/arm64 -o ../build/proxy.xcframework
echo 'Build successful...'