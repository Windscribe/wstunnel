#!/usr/bin/env bash
rm -r build
platforms=("windows/arm64" "windows/amd64" "darwin/amd64" "linux/amd64" "linux/arm64")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name='wstunnel-'$GOOS'-'$GOARCH
	if [ "$GOOS" = "windows" ]; then
		output_name+='.exe'
	fi
	echo "Building $output_name"
	env GOOS="$GOOS" GOARCH="$GOARCH" go build -o build/$output_name -a -gcflags=all="-l -B" -ldflags="-w -s"
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred!'
		exit 1
	fi
done