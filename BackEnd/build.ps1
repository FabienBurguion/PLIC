$env:GOOS="linux"
$env:GOARCH="amd64"
cd http-handler
go build -o ../bootstrap
cd ..
Compress-Archive -Path bootstrap -DestinationPath function.zip -Force