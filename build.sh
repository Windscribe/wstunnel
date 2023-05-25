echo 'Setting go bind path...'
export PATH=$PATH:~/go/bin
echo 'Clean up...'
go mod tidy
mkdir build
cd proxy
echo 'Building...'
gomobile bind -o ../build/proxy.aar  -javapkg com.windscribe