echo 'Setting go bind path...'
export PATH=$PATH:~/go/bin
echo 'Clean up...'
go mod tidy
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
rm -rf ./build
mkdir -p ./build/publications/release
cd proxy
echo 'Building...'
gomobile bind -o ../build/publications/release/proxy.aar  -javapkg com.windscribe
echo 'Build successful.'
