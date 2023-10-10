export PATH=$PATH:~/go/bin
go mod tidy
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
mkdir build
cd proxy
export CGO_ENABLED=1
export CGO_CFLAGS="-fstack-protector-strong"
gomobile bind -o ../build/proxy.aar  -javapkg com.windscribe -ldflags "-s -w"
echo 'Build successful...'