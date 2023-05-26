echo 'Setting go bind path...'
export PATH=$PATH:~/go/bin
echo 'Clean up...'
go mod tidy
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
mkdir build
cd proxy
echo 'Building...'
gomobile bind -o ../build/proxy.aar  -javapkg com.windscribe