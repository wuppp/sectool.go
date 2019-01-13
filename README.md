# sectool.go

## tools
```
portscan (tcp端口扫描)
httpbannerscan (扫描httpbanner)
ping (批量ping)
```

## build
```bash
# GOOS=windows/linux/darwin
# GOARCH=amd64/386 

$ export GOPATH=`pwd`

$ go build portscan.go
$ go build httpbanner.go
```
