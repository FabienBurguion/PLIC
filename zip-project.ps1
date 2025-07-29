$env:GOOS="linux"
$env:GOARCH="amd64"
Set-location http-handler
go build -o ../bootstrap
Set-location ..
Compress-Archive -Path bootstrap -DestinationPath function.zip -Force
Remove-Item bootstrap