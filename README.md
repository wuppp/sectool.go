# sectool.go

## tools
```
portscan (tcp端口扫描)
httpbannerscan (扫描httpbanner)
ping (批量ping)
ssh (ssh爆破)
```

## build
```bash
# GOOS=windows/linux/darwin
# GOARCH=amd64/386 

$ export GOPATH=`pwd`
$ go get -u "golang.org/x/crypto/ssh"

$ go build portscan.go
$ go build httpbanner.go
```
