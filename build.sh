export PATH=$PATH:cfgo/bin
go mod tidy
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
mkdir build
cd proxy
gomobile bind -o ../build/proxy.aar  -javapkg com.windscribe
echo 'Build successful...'